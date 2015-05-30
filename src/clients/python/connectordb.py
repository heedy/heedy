import requests
import json
import time
from urlparse import urljoin
from requests.auth import HTTPBasicAuth
from jsonschema import validate, Draft4Validator
import websocket
import threading
import logging



class AuthenticationError(Exception):
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)
class ServerError(Exception):
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

class DataError(Exception):
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

class ConnectionError(Exception):
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

#The base object upon which the rest is built up
class ConnectorObject(object):
    def __init__(self,connectordb,opath):
        self.db = connectordb
        self.metaname = opath

        self.metadata = None
    def refresh(self):
        #Reload data about user from the server
        self.metadata = self.db.urlget(self.metaname).json()

    @property
    def data(self):
        #Returns the raw dict returned by querying for the device
        if self.metadata is None:
            self.refresh()
        return self.metadata

    def delete(self):
        #Delete the device
        self.db.urldelete(self.metaname)
    @property
    def exists(self):
        #Property is True if object exists, false otherwise
        try:
            self.refresh()
        except:
            return False
        return True

    def set(self,props):
        self.metadata = self.db.urlput(self.metaname,props).json()

    @property
    def name(self):
        return self.data["name"]
    @name.setter
    def name(self,value):
        self.set({"name": value})

class User(ConnectorObject):
    def create(self,email,password):
        #Create the given user
        self.metadata = self.db.urlpost(self.metaname,{"email": email,"password": password}).json()

    def setpassword(self,password):
        self.set({"password": password})

    @property
    def email(self):
        return self.data["email"]
    @email.setter
    def email(self,value):
        self.set({"email": value})

    @property
    def admin(self):
        if not "admin" in self.data:
            return False
        return self.data["admin"]
    @admin.setter
    def admin(self,value):
        self.set({"admin": value})

    def devices(self):
        #Returns the list of users accessible to this operator
        devs = []
        result = self.db.urlget(self.metaname+"/?q=ls")
        for d in result.json():
            tmpd = Device(self.db,d["name"])
            tmpd.metadata = d
            devs.append(tmpd)
        return devs

    def __getitem__(self,val):
        return Device(self.db,self.metaname+"/"+val)

class Device(ConnectorObject):
    def create(self):
        self.metadata = self.db.urlpost(self.metaname).json()

    def streams(self):
        #Returns the list of users accessible to this operator
        strms = []
        result = self.db.urlget(self.metaname+"/?q=ls")
        for s in result.json():
            tmps = Stream(self.db,s["name"])
            tmps.metadata = s
            strms.append(tmps)
        return strms

    def __getitem__(self,val):
        return Stream(self.db,self.metaname+"/"+val)

    @property
    def admin(self):
        if not "admin" in self.data:
            return False
        return self.data["admin"]
    @admin.setter
    def admin(self,value):
        self.set({"admin": value})

    @property
    def nickname(self):
        return self.data["nickname"]
    @nickname.setter
    def nickname(self,value):
        self.set({"nickname": value})

    @property
    def apikey(self):
        return self.data["apikey"]
    @apikey.setter
    def apikey(self,value):
        self.set({"apikey": value})

    def resetKey(self):
        self.apikey=""
        return self.metadata["apikey"]

    @property
    def devicename(self):
        return self.metaname.split("/")[1]

    @property
    def username(self):
        return self.metaname.split("/")[0]

    @property
    def user(self):
        #Gets the device's user
        return User(self.db,self.username)
    @property
    def name(self):
        return self.metaname

class Stream(ConnectorObject):
    def create(self,schema):
        Draft4Validator.check_schema(schema)
        self.metadata = self.db.urlpost(self.metaname,schema).json()

    @property
    def nickname(self):
        return self.data["nickname"]
    @nickname.setter
    def nickname(self,value):
        self.set({"nickname": value})

    @property
    def schema(self):
        return self.data["schema"]

    def __len__(self):
        return int(self.db.urlget(self.metaname+"/length").text)

    def insertMany(self,o):
        self.db.urlupdate(self.metaname,o)

    def insert(self,o):
        self.insertMany([{"t": int(time.time()),"d":o}])

    def __getitem__(self,obj):
        if isinstance(obj,slice):
            return self.db.urlget(self.metaname+"/data?i1="+str(obj.start)+"&i2="+str(obj.stop)).json()
        else:
            return self.db.urlget(self.metaname+"/data?i1="+str(obj)+"&i2="+str(obj+1)).json()[0]
    def __call__(self,t1,t2=0,limit=0):
        return self.db.urlget(self.metaname+"/data?t1="+str(t1)+"&t2="+str(t2)+"&limit="+str(limit)).json()



class ConnectorDB(Device):
    #Connect to ConnectorDB given an user/device name and password/apikey long with an optional url to the server.
    #Alternately, you can log in using your username and password by setting
    #your password to apikey, and name to username.
    def __init__(self,user,password,url="https://connectordb.com"):
        self.auth = HTTPBasicAuth(user,password)
        self.url = url

        
        self.subscriptions = {}
        self.ws = None
        self.ws_thread = None
        self.on_message = None

        Device.__init__(self,self,self.urlget("?q=this").text)


    def __del__(self):
        if self.ws is not None:
            try:
                self.ws.close()
            except:
                pass

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


    def getWebsocketURI(self):
        #Extract the
        ws_url = "wss://" + self.url[8:]
        if self.url.startswith("http://"):   #Unsecured websocket is only really there for testing
            ws_url = "ws://"+self.url[7:]
        return ws_url
    def getHeaders(self):
        class tmpObj():
            def __init__(self):
                self.headers= {}
        tmp = tmpObj()
        self.auth(tmp)
        return tmp.headers

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



    #Does error handling for a request result
    def handleresult(self,r):
        if r.status_code==401 or r.status_code==403:
            raise AuthenticationError(r.text)
        elif r.status_code !=200:
            raise ServerError(r.text)
        return r

    #Direct CRUD requests with the given location and optionally data, which handles authentication and error management
    def urlget(self,location):
        return self.handleresult(requests.get(urljoin(self.url,location),auth=self.auth))
    def urldelete(self,location):
        return self.handleresult(requests.delete(urljoin(self.url,location),auth=self.auth))
    def urlpost(self,location,data={}):
        return self.handleresult(requests.post(urljoin(self.url,location),auth=self.auth,
                                                 headers={'content-type': 'application/json'},data=json.dumps(data)))
    def urlput(self,location,data):
        return self.handleresult(requests.put(urljoin(self.url,location),auth=self.auth,
                                                 headers={'content-type': 'application/json'},data=json.dumps(data)))
    def urlupdate(self,location,data):
        return self.handleresult(requests.request("UPDATE",urljoin(self.url,location),auth=self.auth,
                                                 headers={'content-type': 'application/json'},data=json.dumps(data)))

    def subscribe(self,address=None,callback=None):
        #if address is None:
        #    Device.subscribe(self)
        self.__connectWebsocket()
        self.ws.send(json.dumps({"cmd": "subscribe","arg": address}))

    def getuser(self,usrname):
        return User(self,usrname)

    def users(self):
        #Returns the list of users accessible to this operator
        usrs = []
        for u in self.urlget("?q=ls").json():
            tmpu = self.getuser(u["name"])
            tmpu.metadata = u
            usrs.append(tmpu)
        return usrs

    #We want to be able to get an arbitrary user/device/stream in a simple way
    def __call__(self,address):
        n = address.count("/")
        if n==0:
            return User(self,address)
        elif n==1:
            return Device(self,address)
        else:
            return Stream(self,address)

