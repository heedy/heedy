import websocket
import threading
import logging

class WebsocketHandler(object):
    #SubscriptionHandler manages the websocket connection and fires off all subscriptions
    def __init__(self,url,basicAuth):
        self.uri = self.getWebsocketURI(url)
        self.headers = self.getAuthHeaders(basicAuth)


    def getWebsocketURI(self,url):
        #Given a URL to the REST API, get the websocket uri
        ws_url = "wss://" + self.url[8:]
        if self.url.startswith("http://"):   #Unsecured websocket is only really there for testing
            ws_url = "ws://"+self.url[7:]
        return ws_url

    def getAuthHeaders(self,basicAuth):
        #Use a cheap hack to extract the basic auth header from a requests HTTPBasicAuth object
        class tmpObject():
            def __init__(self):
                self.headers = {}
        tobj = tmpObject()
        basicAuth(tobj)

        headers = []
        for header in tobj.headers:
            headers.append("%s: %s"%(header,tobj.headers[header]))
        return headers

    def subscribe(self,uri,callback):
        #Subscribes to the given 
        pass


"""
def __on_wsmessage(self,ws,msg):
        logging.debug("websocket message: %s",msg)
        if self.on_message is not None:
            self.on_message(json.loads(msg))
    def __on_wsopen(self,opn):
        logging.debug("websocket open")
        self.openlock.release()
    def __on_wsclose(self,arg):
        logging.debug("websocket close")
        self.ws=None
        self.ws_thread=None
        self.openlock.release()
        
    def __on_error(self,e):
        logging.debug("ERROR %s",str(e))
        self.ws=None
        self.ws_thread=None
        self.openlock.release()


    
    

    #Connects to the websocket if there is no connection active
    def __connectWebsocket(self):
        if self.ws_thread is None:
            authheader = self.getHeaders()
            self.ws = websocket.WebSocketApp(self.getWebsocketURI(),header=["Authorization: %s"%(authheader["Authorization"],)],
                                             on_message = self.__on_wsmessage,
                                             on_close = self.__on_wsclose,
                                             on_open = self.__on_wsopen)
            self.ws_thread = threading.Thread(target=self.ws.run_forever)
            self.ws_thread.daemon=True

            self.ws_isrunning = None
            self.openlock = threading.Lock()
            self.openlock.acquire()
            self.ws_thread.start()
            self.openlock.acquire()
            if self.ws is None:
                raise ConnectionError("Could not connect to "+self.getWebsocketURI())
"""