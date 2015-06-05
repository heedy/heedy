"""
The ConnectorDB python client
Made in 2015 by the ConnectorDB team.
"""


import requests
import json

from urlparse import urljoin
from requests.auth import HTTPBasicAuth

from errors import *
from _device import Device
from _user import User
from _stream import Stream
from _websocket import WebsocketHandler


API_URL = "https://connectordb.com/api/v1/"

class ConnectorDB(Device):
    #Connect to ConnectorDB given an user/device name and password/apikey long with an optional url to the server.
    def __init__(self,user,password,url=API_URL):

        if not url.endswith("/"):
            url = url +"/"

        self.auth = HTTPBasicAuth(user,password)
        self.url = url

        self.ws = WebsocketHandler(self.url,self.auth)

        Device.__init__(self,self,self.urlget("?q=this").text)

    def ping(self):
        #Makes sure the connectino is open, and auth is working
        self.urlget("?q=this")

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

