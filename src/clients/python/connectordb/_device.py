from _cobject import ConnectorObject
import _user
import _stream

class Device(ConnectorObject):
    def create(self):
        self.metadata = self.db.urlpost(self.metaname).json()

    def streams(self):
        #Returns the list of streams belonging to the device
        strms = []
        result = self.db.urlget(self.metaname+"/?q=ls")
        for s in result.json():
            tmps = _stream.Stream(self.db,s["name"])
            tmps.metadata = s
            strms.append(tmps)
        return strms

    def __getitem__(self,val):
        #Gets the child stream by name
        return _stream.Stream(self.db,self.metaname+"/"+val)

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
        return _user.User(self.db,self.username)
    @property
    def name(self):
        return self.metaname