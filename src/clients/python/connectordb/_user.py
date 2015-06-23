
from _cobject import ConnectorObject
import _device

class User(ConnectorObject):
    #This object represents a ConnectorDB user.

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
        #Returns the list of devices belonging to the user
        devs = []
        result = self.db.urlget(self.metaname+"/?q=ls")
        for d in result.json():
            tmpd = _device.Device(self.db,self.metaname+"/"+d["name"])
            tmpd.metadata = d
            devs.append(tmpd)
        return devs

    def __getitem__(self,val):
        #Gets a child device by its name
        return _device.Device(self.db,self.metaname+"/"+val)

    def __repr__(self):
        return "[User:%s]"%(self.metaname,)