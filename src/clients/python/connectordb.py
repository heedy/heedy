import requests
import json
from requests.auth import HTTPBasicAuth
import time

"""
NOTE: THIS IS SHIT CODE. BUILT TO JMIFW. Will be un-failed in next version.
"""

class User(object):
    def __init__(self,name,password,url="https://connectordb.com"):
        self.name=name
        self.password=password
        self.url = url+"/api/v1/json/"
    def query(self,q,getval):
        r = requests.get(self.url+q,auth=HTTPBasicAuth(self.name,self.password),headers={'content-type': 'application/json'})
        if r.status_code!=200:
            return None
        return r.json()[getval]
    def users(self):
        return self.query("user/","Users")

    def devices(self):
        return self.query(self.name+"/device/","Devices")
    def addDevice(self,name):
        r = requests.post(self.url+self.name+"/device/",data=json.dumps({"Name": name}),auth=HTTPBasicAuth(self.name,self.password),headers={'content-type': 'application/json'})
        return r.status_code==200
    def __getitem__(self,val):
        devices = self.devices()
        for d in devices:
            if d["Name"]==val:
                return Device(self.name+"/"+str(d["Id"]),d["ApiKey"],dname=d["Name"],url=self.url)
        
        raise Exception("Could not find device")
    def __str__(self):
        return "User("+self.name+")"



class Device(object):
    def __init__(self,name,apikey,dname="",url="https://connectordb.com"):
        self.dname = dname
        self.name = str(name)
        self.apikey=apikey
        self.url=url
        if not self.url.endswith("/api/v1/json/"):
            self.url = url+"/api/v1/json/"
    def query(self,q,getval):
        r = requests.get(self.url+self.name+"/"+q,auth=HTTPBasicAuth("",self.apikey))
        if r.status_code!=200:
            return None
        return r.json()[getval]
    def streams(self):
        return self.query("stream/","Streams")
    
    def metadata(self):
        return self.query("/","Devices")[0]

    def __getitem__(self,val):
        streams = self.streams()
        for d in streams:
            if d["Name"]==val:
                return Stream(self,d)
        
        raise Exception("Could not find stream")
    def addStream(self,name,type="s"):
        r = requests.post(self.url+self.name+"/stream/",json.dumps({'Name':name,"Type":type}),auth=HTTPBasicAuth("",self.apikey),headers={'content-type': 'application/json'})
        return r.status_code==200
    def delete(self):
        r= requests.delete(self.url+self.name+"/",auth=HTTPBasicAuth("",self.apikey))
    def __str__(self):
        return "Device("+self.dname+")"

class Stream(object):
    def __init__(self,parentdevice, streamdict):
        self.dev = parentdevice
        self.stream = streamdict
    def __str__(self):
        return "Stream("+self.stream["Name"]+","+stream["Type"],")"
    def delete(self):
        r= requests.delete(self.dev.url+self.dev.name+"/"+str(self.stream["Id"])+"/",auth=HTTPBasicAuth("",self.dev.apikey))
    def insert(self,point):
        r= requests.post(self.dev.url+self.dev.name+"/"+str(self.stream["Id"])+"/point/",json.dumps({'D':point,'T': str(time.time())}),auth=HTTPBasicAuth("",self.dev.apikey))
        return r.status_code==200
    def getslice(self,stype,start,stop):
        
        result = self.dev.query("/"+str(self.stream["Id"])+"/point/"+stype+"/"+str(start)+"/"+str(stop)+"/","Data")
        timestamps = []
        data = []
        if result is None:
            return ([],[])
        for i in xrange(len(result)):
            timestamps.append(result[i]["T"])
            data.append(result[i]["D"])
        return (timestamps,data)

    def __getitem__(self,i):
        if (isinstance(i,slice)):
            start = i.start
            stop = i.stop
            return self.getslice("i",start,stop)
    def __call__(self,t1,t2):
        t1 = int(t1*1e9)
        t2 = int(t2*1e9)
        return self.getslice("t",t1,t2)


if (__name__=="__main__"):
    import getpass
    usr = User("test","test","http://localhost:8080")
    print usr["AndroidSDKbuiltforx86"]["plugged_in"].insert(False)
    print usr["AndroidSDKbuiltforx86"]["plugged_in"][0:20]
    

