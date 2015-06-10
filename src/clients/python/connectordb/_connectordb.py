"""
The ConnectorDB python client
Made in 2015 by the ConnectorDB team.
"""


import json
import time
from urlparse import urljoin
from requests import Session
from requests.auth import HTTPBasicAuth

from errors import *
from _device import Device
from _user import User
from _stream import Stream
from _websocket import WebsocketHandler


API_URL = "https://connectordb.com"

class ConnectorDB(Device):
    #Connect to ConnectorDB given an user/device name and password/apikey long with an optional url to the server.
    def __init__(self,user,password,url=API_URL):

        if not url.startswith("http"):
            url = "http://"+url

        if not url.endswith("/"):
            url = url +"/"

        self.url = urljoin(url,"/api/v1/")

        auth = HTTPBasicAuth(user,password)
        self.r = Session()  #A Session allows us to reuse connections
        self.r.auth = auth
        self.r.headers.update({'content-type': 'application/json'})


        self.ws = WebsocketHandler(self.url,auth)
        self.__wsinsert = False

        Device.__init__(self,self,self.urlget("?q=this").text)

    def ping(self):
        #Makes sure the connectino is open, and auth is working
        self.urlget("?q=this")

    def handleresult(self,r):
        """Handles HTTP error codes for a given request result

        Raises:
            AuthenticationError on the appropriate 4** errors
            ServerError if the response is not an ok (200)

        Arguments:
            r -- The request result
        """
        if r.status_code in [401, 403]:
            raise AuthenticationError(r.text)
        elif r.status_code !=200:
            raise ServerError(r.text)
        return r

    #Direct CRUD requests with the given location and optionally data, which handles authentication and error management
    def urlget(self,location,cmd="d/"):
        return self.handleresult(self.r.get(urljoin(self.url+cmd,location)))
    def urldelete(self,location,cmd="d/"):
        return self.handleresult(self.r.delete(urljoin(self.url+cmd,location)))
    def urlpost(self,location,data={},cmd="d/"):
        return self.handleresult(self.r.post(urljoin(self.url+cmd,location),data=json.dumps(data)))
    def urlput(self,location,data,cmd="d/"):
        return self.handleresult(self.r.put(urljoin(self.url+cmd,location),data=json.dumps(data)))
    def urlupdate(self,location,data,cmd="d/"):
        return self.handleresult(self.r.request("UPDATE",urljoin(self.url+cmd,location),data=json.dumps(data)))
    
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

    #wsinsert is the property which specifies whether inserts are attempted thru websockets
    @property
    def wsinsert(self):
        #Returns whether or not websocket is used for insert
        return self.__wsinsert
    @wsinsert.setter
    def wsinsert(self,value):
        self.__wsinsert = value
        if value:
            self.wsconnect()

    #Connect and disconnect tell whether to use websocket or not
    def wsconnect(self):
        self.ws.connect()
    def wsdisconnect(self):
        self.ws.disconnect()

    def sleepforever(self):
        #This isn't really anything interesting
        while True:
            time.sleep(100)

    def __repr__(self):
        return "[ConnectorDB:%s]"%(self.metaname,)