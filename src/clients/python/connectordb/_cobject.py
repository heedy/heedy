"""
The ConnectorObject is the underlying object which holds methods which all other objects share
"""

class ConnectorObject(object):
    #The base object upon which the rest is built up.
    def __init__(self,connectordb,opath):
        self.db = connectordb
        self.metaname = opath

        self.metadata = None

    def refresh(self):
        #Refresh reloads the object's metadata from the server
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
        #Attempts to set the given properties of the object
        self.metadata = self.db.urlput(self.metaname,props).json()

    @property
    def name(self):
        #Returns the object's name
        return self.data["name"]
    @name.setter
    def name(self,value):
        self.set({"name": value})

    def subscribe(self,callback):
        #Subscribes to the given object
        self.db.ws.subscribe(self.metaname,callback)

    def unsubscribe(self):
        self.db.ws.unsubscribe(self.metaname)