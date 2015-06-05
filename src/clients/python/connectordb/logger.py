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
    #Allows logging datapoints for a deferred sync with connectordb (allowing to sync eg. once an hour,
    #despite taking data continuously)
    def __init__(self,dbfile,on_create=None):
        #Given a database file, and an optional "on create" callback, open the logger
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

    def __contains__(self,val):
        return val in self.streams

    def __ensureDatabase(self):
        #Run by commands that require a connection to the REST interface to ensure that a connectordb
        #object is connected
        if self.cdb is None:
            if len(self.name)==0 or len(self.apikey)==0 or len(self.url)==0:
                raise errors.ConnectionError("Logger does not have login data set! Can't log in without login data!")
            self.cdb = ConnectorDB(self.name,self.apikey,self.url)

    def __createDatabase(self):
        created = False
        #Create the database tables that will make up the cache if they don't exist yet
        c = self.conn.cursor()
        c.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='cache';")
        if c.fetchone() is None:
            created=True
            logging.debug("Creating table cache")
            c.execute("CREATE TABLE cache (streamname text, timestamp real, data text);")
        c.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='streams';")
        if c.fetchone() is None:
            created=True
            logging.debug("Creating table streams")
            c.execute("CREATE TABLE streams (streamname TEXT PRIMARY KEY, schema TEXT);")
        c.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='metadata';")
        if c.fetchone() is None:
            created=True
            logging.debug("Creating table metadata")
            c.execute("CREATE TABLE metadata (devicename TEXT, apikey TEXT, url TEXT, lastsync real, syncperiod real, userdata TEXT);")
            #The default sync period is 10 minutes
            c.execute("INSERT INTO metadata VALUES ('','',?,0,600,'{}');",(API_URL,))
        return created

    def __clearCDB(self):
        self.synclock.acquire()
        self.cdb = None
        self.synclock.release()

    @property
    def syncperiod(self):
        return self.__syncperiod
    @syncperiod.setter
    def syncperiod(self,value):
        self.__syncperiod = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET syncperiod=?;",(value,))

    @property
    def lastsync(self):
        return self.__lastsync
    @lastsync.setter
    def lastsync(self,value):
        self.__lastsync = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET lastsync=?;",(value,))

    @property
    def name(self):
        return self.__devicename
    @name.setter
    def name(self,value):
        self.__devicename = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET devicename=?;",(value,))
        self.__clearCDB()


    @property
    def apikey(self):
        return self.__apikey
    @apikey.setter
    def apikey(self,value):
        self.__apikey = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET apikey=?;",(value,))
        self.__clearCDB()

    @property
    def url(self):
        return self.__url
    @url.setter
    def url(self,value):
        self.__url = value
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET url=?;",(value,))
        self.__clearCDB()
    
    #The data property allows the user to save settings/data in the database, so that
    #there does not need to be extra code messing around with settings
    @property
    def data(self):
        c = self.conn.cursor()
        c.execute("SELECT userdata FROM metadata;")
        return json.loads(c.fetchone()[0])
    @data.setter
    def data(self,value):
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET userdata=?;",(json.dumps(value),))

    def __loadMeta(self):
        c = self.conn.cursor()
        c.execute("SELECT devicename,apikey,lastsync,syncperiod,url FROM metadata;")
        self.__devicename,self.__apikey,self.__lastsync,self.__syncperiod,self.__url = c.fetchone()

    def __loadStreams(self):
        c = self.conn.cursor()
        c.execute("SELECT * FROM streams;")
        self.streams={}
        for row in c.fetchall():
            self.streams[row[0]]= json.loads(row[1])
    
    def setlogin(self,devicename,apikey,url="https://connectordb.com/api/v1"):
        cdb = ConnectorDB(devicename,apikey,url=url)
        self.synclock.acquire()
        self.cdb = cdb
        self.synclock.release()

        self.name = devicename
        self.apikey = apikey
        self.url = url

    def __len__(self):
        #Returns the number of datapoints currently cached
        c = self.conn.cursor()
        c.execute("SELECT COUNT() FROM cache;");
        return c.fetchone()[0]

    def addStream(self,streampath,schema=None):
        #Adds the given stream to the streams that logger can cache.
        #Requires an active internet connection. If schema is not None, it creates the stream
        #if it does not exist
        self.__ensureDatabase()

        stream = None
        #Allows us to refer to streams either by the full path, or by relative path to the device
        if streampath.count("/")==0:
            stream = self.cdb[streampath]
        else:
            stream = self.cdb(streampath)
        
        if not stream.exists:
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
        try:
            logging.debug("Syncing with connectordb server")
            #Syncs the cache with connectordb
            self.__ensureDatabase()

            #First, let's ping the server to make sure we have internet
            self.cdb.ping()

            #Alright - the sync needs to be locked. Doing 2 syncs at the same time is a BIG no-no
            self.synclock.acquire()

            c = self.conn.cursor()
            for stream in self.streams:
                s = self.cdb(stream)
                if not s.exists:
                    self.synclock.release()
                    DataError("Stream %s no longer exists!"%(stream,))
                c.execute("SELECT * FROM cache WHERE streamname=? ORDER BY timestamp ASC;",(stream,))
                datapointArray=[]
                for dp in c.fetchall():
                    datapointArray.append({"t":dp[1],"d":json.loads(dp[2])})
                if len(datapointArray)>0:
                    logging.debug("%s: syncing %i datapoints"%(stream,len(datapointArray)))
                    try:
                        s.insertMany(datapointArray)
                    except:
                        self.synclock.release()
                        raise

                    #If there was no error inserting, delete the datapoints from the cache
                    c.execute("DELETE FROM cache WHERE streamname=? AND timestamp <=?",(stream,datapointArray[-1]["t"]))

            self.lastsync = time.time()

            self.synclock.release()
        except:
            #Handle the sync failure callback
            reraise = True
            if self.on_syncfail is not None:
                reraise = not self.on_syncfail(self)
            if reraise:
                raise

    def __run(self):
        try:
            self.sync()
        except:
            logging.warn("ConnectorDB sync failed")
        logging.debug("Next sync in "+str(self.syncperiod))
        self.syncer = threading.Timer(self.syncperiod,self.__run)
        self.syncer.start()

    def start(self,period=None):
        logging.debug("Started running background sync with period "+str(self.syncperiod))
        #Runs the syncer in the background with the given sync period
        if period is not None:
            self.syncperiod = period
        
        self.syncer = threading.Timer(self.syncperiod,self.__run)
        self.syncer.start()

    def stop(self):
        #Stops the syncer
        if self.syncer is not None:
            self.syncer.cancel()
            self.syncer= None