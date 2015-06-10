from connectordb.logger import ConnectorLogger
import platform
import threading
import logging

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
        logging.info("Creating cache")
        #Since we are debugging without connectordb access, we manually set the device
        #and force the streams

        c.data = {
            "keypresses": True,
            "activewindow": True,
            "gathertime": 60.0, #60 seconds
            "isrunning": False
            }
        
        logging.info("cache created")

    def setupstreams(self):
        try:
            c = self.cache
            if not "keypresses" in c:
                logging.info("Adding keypress stream")
                c.addStream("keypresses",{"type":"integer",
                                                        "description": "Number of keypresses in the time period of gathering"})
            if not "activewindow" in c:
                logging.info("Adding active window stream")
                c.addStream("activewindow",{"type":"string",
                                                        "description": "The currently active window titlebar text"})
            return ""
        except Exception as e:
            return str(e)

    def gather(self):
        if self.gatherer.log_keypresses:
            kp = self.gatherer.keypresses()
            self.cache.insert("keypresses",kp)
        if self.gatherer.log_activewindow:
            wt = self.gatherer.windowtext()
            self.cache.insert("activewindow",wt)
        logging.info("There are %i datapoints in cache."%(len(self.cache),))

    def __run(self):
        gather_period = self.cache.data["gathertime"]
        self.gather()
        self.syncer = threading.Timer(gather_period,self.__run)
        self.syncer.start()

    def start(self):
        try:
        
            #Set the default options
            d = self.cache.data
            d["isrunning"]=True
            self.cache.data = d
            gather_period = d["gathertime"]
            logging.info("Starting logging with period "+str(gather_period))
            self.gatherer.start()

            if self.syncer is None:
                self.syncer = threading.Timer(gather_period,self.__run)
                self.syncer.start()


            self.cache.start()
            return True
        except:
            return False

    def stop(self):
        logging.info("Stopping gathering")
        self.cache.stop()
        self.gatherer.stop()
        if self.syncer is not None:
            self.syncer.cancel()
            self.syncer= None
        