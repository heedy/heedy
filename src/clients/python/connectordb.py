import requests
import json
from urlparse import urljoin
from requests.auth import HTTPBasicAuth


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



class ConnectorDB(object):
    #Connect to ConnectorDB given an apikey and an optional url to the server.
    #Alternately, you can log in using your username and password by setting
    #your password to apikey, and name to username.
    def __init__(self,apikey,url="https://connectordb.com",username=""):
        self.auth = HTTPBasicAuth(username,apikey)
        self.url = url
        
        #We don't actually know our name - we only have an api key, so let's get the device.
        #This also gives us a chance to make sure our auth is working.
        self.name = self.urlget("d/this").text

        
    #Does error handling for a request result
    def handleresult(self,r):
        if r.status_code==401 or r.status_code==403:
            raise AuthenticationError(r.text)
        elif r.status_code > 400:
            raise ServerError(r.text)
        return r

    #Direct CRUD requests with the given location and optionally data, which handles authentication and error management
    def urlget(self,location):
        return self.handleresult(requests.get(urljoin(self.url,location),auth=self.auth))
    def urldelete(self,location):
        return self.handleresult(requests.delete(urljoin(self.url,location),auth=self.auth))
    def urlpost(self,location,data):
        return self.handleresult(requests.post(urljoin(self.url,location),auth=self.auth,
                                                 headers={'content-type': 'application/json'},data=json.dumps(data)))
    def urlput(self,location,data):
        return self.handleresult(requests.put(urljoin(self.url,location),auth=self.auth,
                                                 headers={'content-type': 'application/json'},data=json.dumps(data)))

    @property
    def username(self):
        return self.name.split("/")[0]

    @property
    def devicename(self):
        return self.name.split("/")[1]

    @property
    def user(self):
        return User(self,self.username)

    

class User(object):
    def __init__(self,connectordb,username):
        self.db = connectordb
        self.name = username

        self.refresh()

    def refresh(self):
        self.metadata = self.db.get(self.name).json()

    def delete(self):
        self.db.urldelete(self.name)
