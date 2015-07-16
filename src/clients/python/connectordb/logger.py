"""
The logger handles all communication with the server, and stores data in a sql database until a sync is ready to happen
"""

import apsw
import logging
import errors
import time
import threading

import json
from jsonschema import validate

from _connectordb import ConnectorDB,API_URL


class ConnectorLogger(object):
    """Allows logging datapoints for a deferred sync with connectordb (allowing to sync eg. once an hour,
    despite taking data continuously)"""
    def __init__(self,dbfile,on_create=None):
        """Given a database file, and an optional "on create" callback, open the logger"""
        self.conn = apsw.Connection(dbfile)
        self.dbfile = dbfile

        self.cdb=None

        run_createcallback = self.__createDatabase()
        self.__loadMeta()
        self.__loadStreams()

        self.synclock = threading.Lock()

        self.syncer = None

        if run_createcallback and on_create is not None:
            on_create(self)

        #Set up the callbacks
        self.on_syncfail = None #Called when a sync fails. Returns True if want to stop the error from propagating on

    def __del__(self):
        self.stop()

    def __ensureDatabase(self):
        """Run by commands that require a connection to the REST interface to ensure that a connectordb
        object is connected"""
        if self.cdb is None:
            if len(self.name)==0 or len(self.apikey)==0 or len(self.url)==0:
                raise errors.ConnectionError("Logger does not have login data set! Can't log in without login data!")
            self.cdb = ConnectorDB(self.name,self.apikey,self.url)


    def __createDatabase(self):
        """Create the database tables that will make up the cache if they don't exist yet"""
        c = self.conn.cursor()
        try:
            logging.debug("Creating table cache if not exists")
            c.execute("CREATE TABLE IF NOT EXISTS cache (streamname text, timestamp real, data text);")

            logging.debug("Creating table streams if not exists")
            c.execute("CREATE TABLE IF NOT EXISTS streams (streamname TEXT PRIMARY KEY, schema TEXT);")

            logging.debug("Creating table metadata if not exists")
            c.execute("CREATE TABLE IF NOT EXISTS metadata (devicename TEXT, apikey TEXT, url TEXT, lastsync real, syncperiod real, userdata TEXT);")

            #The default sync period is 10 minutes
            c.execute("INSERT INTO metadata VALUES ('','',?,0,600,'{}');",(API_URL,))
        except apsw.SQLError:
            return False

        return True

    def __clearCDB(self):
        with self.synclock:
            self.cdb = None

    @property
    def syncperiod(self):
        """syncperiod is the time in seconds to wait between attempting to sync data to the ConnectorDB server"""
        return self.__syncperiod
    @syncperiod.setter
    def syncperiod(self,value):
        resync = False
        with self.synclock:
            self.__syncperiod = value
            c = self.conn.cursor()
            c.execute("UPDATE metadata SET syncperiod=?;",(value,))
            if self.syncer is not None:
                resync = True
        if resync:
            self.__setsync()

    @property
    def lastsync(self):
        """lastsync is the timestamp of the most recent synchronization to the server"""
        return self.__lastsync
    @lastsync.setter
    def lastsync(self,value):
        self.__lastsync = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET lastsync=?;",(value,))

    @property
    def name(self):
        """The device name that the logger operates as"""
        return self.__devicename
    @name.setter
    def name(self,value):
        self.__devicename = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET devicename=?;",(value,))
        self.__clearCDB()


    @property
    def apikey(self):
        """The api key that the logger uses to connect to ConnectorDB"""
        return self.__apikey
    @apikey.setter
    def apikey(self,value):
        self.__apikey = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET apikey=?;",(value,))
        self.__clearCDB()

    @property
    def url(self):
        """The URL of the ConnectorDB server"""
        return self.__url
    @url.setter
    def url(self,value):
        self.__url = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET url=?;",(value,))
        self.__clearCDB()

    
    @property
    def data(self):
        """The data property allows the user to save settings/data in the database, so that
            there does not need to be extra code messing around with settings.

        Use this property to save things that can be converted to JSON inside the logger database,
        so that you don't have to mess with configuration files or saving setting otherwise.
        """
        c = self.conn.cursor()
        c.execute("SELECT userdata FROM metadata;")
        return json.loads(c.next()[0])
    @data.setter
    def data(self,value):
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET userdata=?;",(json.dumps(value),))

    def __loadMeta(self):
        c = self.conn.cursor()
        c.execute("SELECT devicename,apikey,lastsync,syncperiod,url FROM metadata;")
        self.__devicename,self.__apikey,self.__lastsync,self.__syncperiod,self.__url = c.next()

    def __loadStreams(self):
        c = self.conn.cursor()
        c.execute("SELECT * FROM streams;")
        self.streams={}
        for row in c.fetchall():
            self.streams[row[0]]= json.loads(row[1])

    def setlogin(self,devicename,apikey,url="https://connectordb.com/api/v1"):
        """Set the login credentials that the logger will use to connect ot ConnectorDB"""
        cdb = ConnectorDB(devicename,apikey,url=url)
        self.synclock.acquire()
        self.cdb = cdb
        self.synclock.release()

        self.name = devicename
        self.apikey = apikey
        self.url = url

    def __len__(self):
        """Returns the number of datapoints currently cached"""
        c = self.conn.cursor()
        c.execute("SELECT COUNT() FROM cache;");
        return c.next()[0]

    def addStream(self,streampath,schema=None):
        """Adds the given stream to the streams that logger can cache.
        Requires an active internet connection. If schema is not None, it creates the stream
        if it does not exist"""
        self.__ensureDatabase()

        stream = None
        #Allows us to refer to streams either by the full path, or by relative path to the device
        if streampath.count("/")==0:
            stream = self.cdb[streampath]
        else:
            stream = self.cdb(streampath)

        if not stream.exists():
            if schema is not None:
                stream.create(schema)
            else:
                raise errors.ServerError("The stream '%s' was not found"%(stream.metaname,))

        self.force_addStream(stream.metaname,stream.schema)

    def force_addStream(self,streampath,streamschema):
        #Forces a stream add without checking connectordb to make sure the stream exists.
        #Requires the full path to the stream
        c = self.conn.cursor()
        c.execute("INSERT OR REPLACE INTO streams VALUES (?,?);",(streampath,json.dumps(streamschema)))

        self.streams[streampath] = streamschema


    def insert(self,streamname,value):
        """Insert into a stream that the logger can cache"""
        if streamname.count("/")==0:
            streamname = self.name + "/"+streamname
        if streamname.count("/")<=1 or not streamname in self.streams:
            raise errors.DataError("Stream not found '%s'"%(streamname,))

        #Validate the schema
        validate(value,self.streams[streamname])

        #Alright - all is validated. Now insert into cache
        logging.debug("CACHE: %s <= %s"%(streamname,value))
        c = self.conn.cursor()
        c.execute("INSERT INTO cache VALUES (?,?,?);",(streamname,time.time(),json.dumps(value)))

    def sync(self):
        """Attempt to sync with the ConnectorDB server"""
        try:
            logging.debug("Syncing with connectordb server")
            #Syncs the cache with connectordb
            self.__ensureDatabase()

            #First, let's ping the server to make sure we have internet
            self.cdb.ping()

            #Alright - the sync needs to be locked. Doing 2 syncs at the same time is a BIG no-no
            with self.synclock:
                c = self.conn.cursor()
                for stream in self.streams:
                    s = self.cdb(stream)
                    if not s.exists():
                        raise errors.DataError("Stream %s no longer exists!"%(stream,))
                    c.execute("SELECT * FROM cache WHERE streamname=? ORDER BY timestamp ASC;",(stream,))
                    datapointArray=[]
                    for dp in c.fetchall():
                        datapointArray.append({"t":dp[1],"d":json.loads(dp[2])})
                    if len(datapointArray)>0:
                        logging.debug("%s: syncing %i datapoints"%(stream,len(datapointArray)))
                        s.insertMany(datapointArray)

                        #If there was no error inserting, delete the datapoints from the cache
                        c.execute("DELETE FROM cache WHERE streamname=? AND timestamp <=?",(stream,datapointArray[-1]["t"]))

                self.lastsync = time.time()

        except:
            #Handle the sync failure callback
            reraise = True
            if self.on_syncfail is not None:
                reraise = not self.on_syncfail(self)
            if reraise:
                raise

    def __setsync(self):
        with self.synclock:
            logging.debug("Next sync attempt in "+str(self.syncperiod))
            if self.syncer is not None:
                self.syncer.cancel()
            self.syncer = threading.Timer(self.syncperiod,self.__run)
            self.syncer.start()

    def __run(self):
        try:
            self.sync()
        except Exception as e:
            logging.warn("ConnectorDB sync failed: "+str(e))
        self.__setsync()

    def start(self,period=None):
        """Starts the logger synchronization service in the background. This allows you to not
        need to worry about syncing with ConnectorDB - you jsut add to cache, and the cache will be syncd
        periodically (accoridng to syncperiod)"""
        self.__ensureDatabase()
        with self.synclock:
            if self.syncer is not None:
                logging.warn("start called on syncer that is already running")
                return

            logging.debug("Started running background sync with period "+str(self.syncperiod))
            #Runs the syncer in the background with the given sync period
            if period is not None:
                self.syncperiod = period

            self.syncer = threading.Timer(self.syncperiod,self.__run)
            self.syncer.start()

    def stop(self):
        """Stops the syncer"""
        with self.synclock:
            if self.syncer is not None:
                self.syncer.cancel()
                self.syncer= None

    def __contains__(self,streampath):
        """Whether the logger is caching the given stream name"""
        if streampath.count("/")==0:
            return self.name+"/"+streampath in self.streams
        else:
            return streampath in self.streams
