from jsonschema import validate, Draft4Validator
import time
from _cobject import ConnectorObject

class Stream(ConnectorObject):
    def create(self,schema):
        #Given a dict representing the stream's JSON Schema, creates the stream
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
        self.insertMany([{"t": time.time(),"d":o}])

    def __getitem__(self,obj):
        #Allows to access the stream's elements as if they were an array
        if isinstance(obj,slice):
            return self.db.urlget(self.metaname+"/data?i1="+str(obj.start)+"&i2="+str(obj.stop)).json()
        else:
            return self.db.urlget(self.metaname+"/data?i1="+str(obj)+"&i2="+str(obj+1)).json()[0]

    def __call__(self,t1,t2=0,limit=0):
        #Allows to get the datapoints in a stream between two times or with a limit
        return self.db.urlget(self.metaname+"/data?t1="+str(t1)+"&t2="+str(t2)+"&limit="+str(limit)).json()

    def subscribe(self,callback,downlink=False,substream=""):
        #Stream subscription is a bit more comples, since a stream can be a downlink and can have substreams
        #so we subscribe according to that
        sname = self.metaname
        if downlink:
            sname += "/downlink"
        elif len(substream)>0:
            sname += "/"+substream
        self.db.ws.subscribe(sname,callback)