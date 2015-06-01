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

        self.gather_period = 60 #One minute
        self.syncer = None

    def create_callback(self,c):
        print "Creating cache"
        #Since we are debugging without connectordb access, we manually set the device
        #and force the streams
        c.setDeviceName("test/test")
        c.force_addStream("test/test/keypresses",{"type":"integer",
                                                  "description": "Number of keypresses in the time period of gathering"})
        c.force_addStream("test/test/activewindow",{"type":"string",
                                                  "description": "The currently active window titlebar text"})
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
        self.gather()
        self.syncer = threading.Timer(self.gather_period,self.__run)
        self.syncer.start()

    def start(self):
        print "Starting logging with period "+str(self.gather_period)
        self.gatherer.start()

        if self.syncer is None:
            self.syncer = threading.Timer(self.gather_period,self.__run)
            self.syncer.start()

    def stop(self):
        print "Stopping gathering"
        self.gatherer.stop()
        if self.syncer is not None:
            self.syncer.cancel()
            self.syncer= None
        