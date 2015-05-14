import requests
import json
from urlparse import urljoin
from requests.auth import HTTPBasicAuth
from jsonschema import validate, Draft4Validator


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
        result = self.db.urlget(self.metaname+"/?special=ls")
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
        result = self.db.urlget(self.metaname+"/?special=ls")
        print result
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


class ConnectorDB(Device):
    #Connect to ConnectorDB given an user/device name and password/apikey long with an optional url to the server.
    #Alternately, you can log in using your username and password by setting
    #your password to apikey, and name to username.
    def __init__(self,user,password,url="https://connectordb.com"):
        self.auth = HTTPBasicAuth(user,password)
        self.url = url

        Device.__init__(self,self,self.urlget("?special=this").text)


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

    def getuser(self,usrname):
        return User(self,usrname)

    def users(self):
        #Returns the list of users accessible to this operator
        usrs = []
        for u in self.urlget("?special=ls").json():
            tmpu = self.getuser(u["name"])
            tmpu.metadata = u
            usrs.append(tmpu)
        return usrs
