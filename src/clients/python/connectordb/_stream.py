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
    def downlink(self):
        v = self.data["downlink"]
        if v is None:
            return False
        return v
        
    @downlink.setter
    def downlink(self,value):
        self.set({"downlink": value})

    @property
    def ephemeral(self):
        v = self.data["ephemeral"]
        if v is None:
            return False
        return v
        
    @ephemeral.setter
    def ephemeral(self,value):
        self.set({"ephemeral": value})

    @property
    def schema(self):
        return self.data["schema"]

    def __len__(self):
        return int(self.db.urlget(self.metaname+"/data?q=length").text)

    def insertMany(self,o,restamp=False):
        #attempt to use websocket if websocket inserts are enabled, but fall back on update if fail
        if self.db.wsinsert:
            if self.db.ws.insert(self.metaname,o):
                return
        if restamp:
            self.db.urlput(self.metaname+"/data",o)
        else:
            self.db.urlpost(self.metaname+"/data",o)

    def insert(self,o):
        self.insertMany([{"d":o}],restamp=True)

    def __getitem__(self,obj):
        #Allows to access the stream's elements as if they were an array
        if isinstance(obj,slice):
            start = obj.start
            if start is None:
                start = 0
            stop = obj.stop
            if stop is None:
                stop = 0
            return self.db.urlget(self.metaname+"/data?i1="+str(start)+"&i2="+str(stop)).json()
        else:
            return self.db.urlget(self.metaname+"/data?i1="+str(obj)+"&i2="+str(obj+1)).json()[0]

    def __call__(self,t1,t2=0,limit=0):
        #Allows to get the datapoints in a stream between two times or with a limit
        return self.db.urlget(self.metaname+"/data?t1="+str(t1)+"&t2="+str(t2)+"&limit="+str(limit)).json()

    def subscribe(self,callback,downlink=False):
        '''Stream subscription is a bit more comples, since a stream can be a downlink and can have substreams
        so we subscribe according to that
        '''
        
        sname = self.metaname
        if downlink:
            sname += "/downlink"
        self.db.ws.subscribe(sname,callback)
        
    def unsubscribe(self,downlink=False):
        sname = self.metaname
        if downlink:
            sname += "/downlink"
        self.db.ws.unsubscribe(sname)

    def __repr__(self):
        return "[Stream:%s]"%(self.metaname,)