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


class ConnectorLogger(object):
    #Allows logging datapoints for a deferred sync with connectordb (allowing to sync eg. once an hour,
    #despite taking data continuously)
    def __init__(self,dbfile,cdb=None):
        #Given a database file and the connectordb database (db is optional, since might want
        #to log without internet
        self.conn = apsw.Connection(dbfile)
        self.dbfile = dbfile

        self.__createDatabase()
        self.__loadMeta()
        self.__loadStreams()

        self.connect(cdb)

        self.synclock = threading.Lock()

        self.syncer = None

    def __del__(self):
        self.stop()

    def __ensureDatabase(self):
        #Run by commands that require a connection to the REST interface to ensure that a connectordb
        #object is connected
        if self.cdb is None:
            raise ConnectionError("The logger does not have a connectordb connection active!")

    def __createDatabase(self):
        #Create the database tables that will make up the cache if they don't exist yet
        c = self.conn.cursor()
        c.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='cache';")
        if c.fetchone() is None:
            logging.debug("Creating table cache")
            c.execute("CREATE TABLE cache (streamname text, timestamp real, data text);")
        c.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='streams';")
        if c.fetchone() is None:
            logging.debug("Creating table streams")
            c.execute("CREATE TABLE streams (streamname TEXT PRIMARY KEY, schema TEXT);")
        c.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='metadata';")
        if c.fetchone() is None:
            logging.debug("Creating table streams")
            c.execute("CREATE TABLE metadata (devicename TEXT, lastsync real, syncperiod real);")
            #The default sync period is 10 minutes
            c.execute("INSERT INTO metadata VALUES ('',0,600);")

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

    def __loadMeta(self):
        c = self.conn.cursor()
        c.execute("SELECT * FROM metadata;")
        self.devicename,self.__lastsync,self.__syncperiod = c.fetchone()

    def __loadStreams(self):
        c = self.conn.cursor()
        c.execute("SELECT * FROM streams;")
        self.streams={}
        for row in c.fetchall():
            self.streams[row[0]]= json.loads(row[1])

    def __setDeviceName(self,devname):
        self.devicename = devname
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET devicename=?;",(devname,))
    
    def connect(self,cdb):
        if cdb is not None:
            #Connects to the given connectorDB database
            self.cdb = cdb

            if self.devicename=="":
                self.__setDeviceName(self.cdb.metaname)

    def __len__(self):
        #Returns the number of datapoints currently cached
        c = self.conn.cursor()
        c.execute("SELECT COUNT() FROM cache;");
        return c.fetchone()[0]

    def addStream(self,streampath):
        #Adds the given stream to the streams that logger can cache.
        #Requires an active internet connection
        self.__ensureDatabase()

        stream = None
        #Allows us to refer to streams either by the full path, or by relative path to the device
        if streampath.count("/")==0:
            stream = self.cdb[streampath]
        else:
            stream = self.cdb(streampath)
        
        if not stream.exists:
            raise errors.ServerError("The stream '%s' was not found"%(stream.metaname,))

        c = self.conn.cursor()
        c.execute("INSERT OR REPLACE INTO streams VALUES (?,?);",(stream.metaname,json.dumps(stream.schema)))

        self.streams[stream.metaname] = stream.schema


    def insert(self,streamname,value):
        if streamname.count("/")==0:
            streamname = self.devicename + "/"+streamname
        if streamname.count("/")<=1 or not streamname in self.streams:
            raise errors.DataError("Stream not found '%s'"%(streamname,))

        #Validate the schema
        validate(value,self.streams[streamname])

        #Alright - all is validated. Now insert into cache
        logging.debug("CACHE: %s <= %s"%(streamname,value))
        c = self.conn.cursor()
        c.execute("INSERT INTO cache VALUES (?,?,?);",(streamname,time.time(),json.dumps(value)))

    def sync(self):
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

    def __run(self):
        self.sync()
        self.syncer = threading.Timer(self.syncperiod,self.__run)
        self.syncer.start()

    def run(self,period=None):
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