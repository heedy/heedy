import websocket
import threading
import logging
import json
import errors
import time
import random

MAX_RECONNECT_TIME_SECONDS = 8 * 60.0
RECONNECT_TIME_BACKOFF_RATE = 1.5
INITIAL_RECONNECT_DELAY = 1.0

class WebsocketHandler(object):
    """WebsocketHandler handles websocket connections to a ConnectorDB server. It gracefully handles
    subscribing, unsubscribing, and deals with dropped connections."""
    
    def __init__(self,url,basicAuth):
        """Given a url, and a object that returns basic auth header when called (as the requests basic auth obj does"""
        self.uri = self.getWebsocketURI(url)
        self.headers = self.getAuthHeaders(basicAuth)

        self.subscriptions = {}
        self.subscription_lock = threading.Lock()

        self.isconnected = False


        self.ws = None
        self.ws_thread = None
        self.ws_openlock = threading.Lock()    #Allows us to synchronously wait for connection to be ready
        self.ws_sendlock = threading.Lock()

        #If it wants a connection, then if websocket dies, it is reconnected immediately
        self.wantsconnection = False
        self.isretry = False
        self.reconnectbackoff= INITIAL_RECONNECT_DELAY
        self.connectedtime = 0.0
        self.lastdisconnect = 0.0

        #The timestamp of the most recent ping message from server
        self.lastping = 0.0
        self.ping_timeout = 60*2 #Set timeout to be 2 minutes without ping
        self.pingtimer = None

    def __del__(self):
        self.disconnect()
        #A weird error was showing up if the program exited without stopping the ping - the timer ran once the time module was unloaded
        if self.pingtimer is not None:
            self.pingtimer.cancel()

    def getWebsocketURI(self,url):
        """Given a URL to the REST API, get the websocket uri"""
        ws_url = "wss://" + url[8:]
        if url.startswith("http://"):   #Unsecured websocket is only really there for testing
            ws_url = "ws://"+ url[7:]
        return ws_url

    def getAuthHeaders(self,basicAuth):
        """Use a cheap hack to extract the basic auth header from a requests HTTPBasicAuth object"""
        class tmpObject():
            def __init__(self):
                self.headers = {}
        tobj = tmpObject()
        basicAuth(tobj)

        headers = []
        for header in tobj.headers:
            headers.append("%s: %s"%(header,tobj.headers[header]))
        return headers

    def unlockopen(self,isconnected=False):
        """Unlocks the open mutex - and returns false if the mutex is already open"""
        self.isconnected = isconnected

        if self.pingtimer is not None:
            self.pingtimer.cancel()
            self.pingtimer = None

        try:
            self.ws_openlock.release()
            return True
        except threading.ThreadError:
            return False


    def __on_message(self,ws,msg):
        """called when a message is received from the connectordb server"""
        msg= json.loads(msg)
        logging.debug("ConnectorDB: Got message for '%s'",msg["stream"])

        #Alright - now that we have the data, we call callbacks, in order:
        self.subscription_lock.acquire()

        def runfnc(fnc):
            #Run the callbacks, but release the lock, just in case subscribing/unsubscribing happens
            #in the callbacks
            self.subscription_lock.release()
            s = msg["stream"]
            res = fnc(s,msg["data"])
            if res==True:
                #This is a convenience function - if True is returned by the subcription callback,
                #it means that the callback wants the same datapoints (without changed timestamps) to be inserted.
                res = msg["data"]
            if res!=False and res is not None and s.endswith("/downlink") and s.count("/")==3:
                #The downlink was acknowledged - write the datapoints through websocket,
                #so that it is visible that it was processed
                self.insert(msg["stream"][:-9],res)
            self.subscription_lock.acquire()

        if msg["stream"] in self.subscriptions:
            runfnc(self.subscriptions[msg["stream"]])


        #Now handle the more general subscriptions to device or user
        pathparts = msg["stream"].split("/")

        if len(pathparts)==3:
            #We don't want to get downlinks or substreams in this
            if pathparts[0] in self.subscriptions:
                runfnc(self.subscriptions[pathparts[0]])

            if pathparts[0]+"/"+pathparts[1] in self.subscriptions:
                runfnc(self.subscriptions[pathparts[0]+"/"+pathparts[1]])

        self.subscription_lock.release()

    def __on_ping(self,ws,data):
        """The server sends ping messages - to ensure that we don't lose connection, we memorize the
        time that the msot recent ping was received - and we check it in __ensure_ping."""
        self.lastping = time.time()


    def __ensure_ping(self):
        """We memorize the timestamp each time we receive a ping from the server. To ensure that the connection
        is actually still alive, this function is run periodically, and if there was no ping within a prespecified
        wait time, then the websocket is closed and a reconnect is attempted"""
        if time is None:    #A weird little bug - on exiting this was sometimes getting an error
            return
        if (time.time()-self.lastping > self.ping_timeout):
            logging.warn("Websocket ping timer timed out!")
            if self.ws is not None:
                self.ws.close()
        self.pingtimer = threading.Timer(self.ping_timeout,self.__ensure_ping)
        self.pingtimer.setDaemon(True)
        self.pingtimer.start()


    def __on_open(self,ws):
        """Called when the websocket is opened"""
        logging.debug("ConnectorDB: Websocket opened")
        self.connectedtime = time.time()
        self.unlockopen(True)

        self.lastping = time.time() #Set the ping timer to current time.
        self.__ensure_ping()



    def __on_close(self,ws):
        """Called when the websocket is closed"""
        if self.wantsconnection and not self.isretry:
            self.__on_error(ws,Exception("Websocket was closed despite wanting a connection..."))
            return
        logging.debug("ConnectorDB: Websocket Closed")
        self.unlockopen()

    def __on_error(self,ws,e):
        """Called when a websocket has an error AND when the websocket is closed despite wanting a connection.
        If this error corresponds to an existing websocket dying, then attempt to reconnect with a backoff"""
        logging.debug("ConnectorDB: Websocket error: %s",str(e))
        v = self.unlockopen()
        if not v or self.isretry:
            if not self.isconnected and self.wantsconnection:
                self.isretry = True

                if self.connectedtime > self.lastdisconnect:
                    self.lastdisconnect = time.time()
                    #We JUST got disconnected. So we reset the backoff parameter based on how
                    #long we were connected
                    if self.lastdisconnect - self.connectedtime > 15*60:
                        self.reconnectbackoff = INITIAL_RECONNECT_DELAY
                    else:
                        self.reconnectbackoff /= RECONNECT_TIME_BACKOFF_RATE #It will be multiplied by it again in a moment


                if self.reconnectbackoff < INITIAL_RECONNECT_DELAY:
                    self.reconnectbackoff = INITIAL_RECONNECT_DELAY
                logging.warn("Disconnected from websocket. Retrying in %.2fs"%(self.reconnectbackoff,))
                #The connection was already unlocked, is not connected, and it wants a connection.
                reconnector = threading.Timer(self.reconnectbackoff,self.__reconnect_callback)
                reconnector.daemon=True
                reconnector.start()

    def __reconnect_callback(self):
        """ Updates the reconnectbackoff in a method similar to TCP Tahoe and
        attempts a reconnect.

        """
        # Double the reconnect time
        self.reconnectbackoff *= RECONNECT_TIME_BACKOFF_RATE

        # don't overflow our backoff time, or else the user will be mad
        if self.reconnectbackoff > MAX_RECONNECT_TIME_SECONDS:
            self.reconnectbackoff = MAX_RECONNECT_TIME_SECONDS

        #Now add randomness to the backoff rate - necessary not to pound the server if it goes down
        self.reconnectbackoff *= 1 + random.uniform(-0.2,0.2)



        try:
            logging.debug("Reconnecting websocket...")
            self.connect(forceretry=True)
            self.__resubscribe()
            logging.warn("Reconnect Successful")
        except:
            pass

    def __resubscribe(self):
        """Subscribe to all existing subscriptions (happens on reconnect)"""
        with self.subscription_lock:
            for sub in self.subscriptions:
                logging.debug("Resubscribing to %s",sub)
                self.send({"cmd": "subscribe", "arg":sub})

    def send(self,cmd):
        """Send the given command thru websocket"""
        with self.ws_sendlock:
            self.ws.send(json.dumps(cmd))

    def insert(self,uri,data):
        """insert the given data to the given stream"""
        if not self.connect():
            return False
        try:
            logging.debug("Inserting thru websocket")
            self.send({"cmd": "insert", "arg": uri,"d": data})
        except:
            return False
        return True

    def subscribe(self,uri,callback):
        """Given a uri to subscribe to, and a callback function, sets up the callback"""
        if not self.connect():
            return False

        logging.debug("Subscribing to %s",uri)
        #Subscribes to the given uri with the given callback
        self.send({"cmd": "subscribe", "arg": uri})
        with self.subscription_lock:
            self.subscriptions[uri] = callback

        return True

    def unsubscribe(self,uri):
        """Unsubscribe from the given uri"""
        logging.debug("Unsubscribing from %s",uri)
        try:
            self.send({"cmd": "unsubscribe", "arg": uri})
        except:
            pass
        #Unsubscribes from the given uri
        with self.subscription_lock:
            del self.subscriptions[uri]

    def disconnect(self):
        """Disconnects the websocket"""
        self.wantsconnection = False
        #Closes the connection if it exists
        if self.ws is not None:
            self.ws.close()


        with self.subscription_lock:
            self.subscriptions = {}

    def connect(self,forceretry=False):
        """Attempts to connect to the websocket - returns False if it is already attempting connection
        and true if the websocket is attempting to connect in background"""
        if not self.isconnected and self.wantsconnection and not forceretry:
            return False    #Means that is in process of retrying
        self.wantsconnection = True
        #Connects to the server if there is no connection active
        if not self.isconnected or forceretry:
            self.ws = websocket.WebSocketApp(self.uri,header=self.headers,
                                             on_message = self.__on_message,
                                             on_close = self.__on_close,
                                             on_open = self.__on_open,
                                             on_error = self.__on_error,
                                             on_ping = self.__on_ping)

            self.ws_thread = threading.Thread(target=self.ws.run_forever)
            self.ws_thread.daemon=True

            self.ws_openlock.acquire()
            self.ws_thread.start()

            #The lock will be released once there is news from the connection
            #so we acquire and release it again
            self.ws_openlock.acquire()
            self.ws_openlock.release()

            if not self.isconnected:
                raise errors.ConnectionError("Could not connect to "+self.uri)
            else:
                self.isretry=False
        return True
