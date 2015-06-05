from connectordb.logger import ConnectorLogger
import platform
import threading

my_os = platform.system()
if my_os=="Windows":
    from winlog import DataGatherer
else:
    #from linlog import DataGatherer
    raise Exception("THIS OPERATING SYSTEM DOES NOT YET HAVE DATA GATHERING IMPLEMENTED")


class DataCache():
    def __init__(self):
        self.cache = ConnectorLogger("cache.db",on_create=self.create_callback)


        self.gatherer = DataGatherer()

        self.syncer = None

        if self.cache.data["isrunning"]:
            self.start()

    def create_callback(self,c):
        print "Creating cache"
        #Since we are debugging without connectordb access, we manually set the device
        #and force the streams
        c.name = "test/test"
        c.force_addStream("test/test/keypresses",{"type":"integer",
                                                  "description": "Number of keypresses in the time period of gathering"})
        c.force_addStream("test/test/activewindow",{"type":"string",
                                                  "description": "The currently active window titlebar text"})
        
        #Set the default options
        c.data = {
            "keypresses": True,
            "activewindow": True,
            "gathertime": 60.0, #60 seconds
            "isrunning": False
            }
        
        print "cache created"

    def gather(self):
        if self.gatherer.log_keypresses:
            kp = self.gatherer.keypresses()
            print "Keypresses:",kp
            self.cache.insert("keypresses",kp)
        if self.gatherer.log_activewindow:
            wt = self.gatherer.windowtext()
            print "WindowText:",wt
            self.cache.insert("activewindow",wt)
        print "There are %i datapoints in cache."%(len(self.cache),)

    def __run(self):
        gather_period = self.cache.data["gathertime"]
        self.gather()
        self.syncer = threading.Timer(gather_period,self.__run)
        self.syncer.start()

    def start(self):
        try:

            #Try to set up the streams (if they exist, then great). If not, then start fails
            self.cache.addStream("keypresses",{"type":"integer",
                                                  "description": "Number of keypresses in the time period of gathering"})
            self.cache.addStream("activewindow",{"type":"string",
                                                  "description": "The currently active window titlebar text"})
        
            #Set the default options
            d = self.cache.data
            d["isrunning"]=True
            self.cache.data = d
            gather_period = d["gathertime"]
            print "Starting logging with period "+str(gather_period)
            self.gatherer.start()

            if self.syncer is None:
                self.syncer = threading.Timer(gather_period,self.__run)
                self.syncer.start()


            self.cache.start()
            return True
        except:
            return False

    def stop(self):
        print "Stopping gathering"
        self.cache.stop()
        self.gatherer.stop()
        if self.syncer is not None:
            self.syncer.cancel()
            self.syncer= None
        