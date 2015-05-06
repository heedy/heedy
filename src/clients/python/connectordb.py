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
    #Connect to ConnectorDB given an user/device name and password/apikey long with an optional url to the server.
    #Alternately, you can log in using your username and password by setting
    #your password to apikey, and name to username.
    def __init__(self,user,password,url="https://connectordb.com"):
        self.auth = HTTPBasicAuth(user,password)
        self.url = url
        
        #We don't actually know our name - we only have an api key, so let's get the device.
        #This also gives us a chance to make sure our auth is working.
        self.name = self.urlget("this").text

        
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
    def thisuser(self):
        return self.user(self.username)

    def user(self,usrname):
        return User(self,usrname)

    def users(self):
        #Returns the list of users accessible to this operator
        usrs = []
        for u in self.urlget("ls").json():
            tmpu = self.user(u["name"])
            tmpu.metadata = u
            usrs.append(tmpu)
        return usrs


    

class User(object):
    def __init__(self,connectordb,username):
        self.db = connectordb
        self.__name = username

        self.metadata = None

    def refresh(self):
        #Reload data about user from the server
        self.metadata = self.db.urlget(self.__name).json()

    @property
    def data(self):
        #Returns the raw dict returned by querying for the user
        if self.metadata is None:
            self.refresh()
        return self.metadata

    def delete(self):
        #Delete the user
        self.db.urldelete(self.__name)

    def create(self,email,password):
        #Create the given user
        self.metadata = self.db.urlpost(self.__name,{"email": email,"password": password}).json()

    @property
    def exists(self):
        #Property is True if user exists, false otherwise
        try:
            self.refresh()
        except:
            return False
        return True

    def set(self,props):
        self.metadata = self.db.urlput(self.__name,props).json()

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


