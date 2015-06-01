"""
The logger handles all communication with the server, and stores data in a sql database until a sync is ready to happen
"""

import sqlite3
import logging
import errors
import time

import json
from jsonschema import validate


class ConnectorLogger(object):
    #Allows logging datapoints for a deferred sync with connectordb (allowing to sync eg. once an hour,
    #despite taking data continuously)
    def __init__(self,dbfile,cdb=None):
        #Given a database file and the connectordb database (db is optional, since might want
        #to log without internet
        self.conn = sqlite3.connect(dbfile)

        self.__createDatabase()
        self.__loadMeta()
        self.__loadStreams()

        self.connect(cdb)

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
            c.execute("INSERT INTO metadata VALUES ('',0,0);")

        self.conn.commit()

    def __loadMeta(self):
        c = self.conn.cursor()
        c.execute("SELECT * FROM metadata;")
        self.devicename,self.lastsync,self.__syncperiod = c.fetchone()

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
        c.commit()
    def __setLastSync(self,lastsync):
        self.lastsync = lastsync
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET lastsync=?;",(lastsync,))
        c.commit()
    def __setSyncPeriod(self,speriod):
        self.__syncperiod = speriod
        c = self.conn.cursor()
        c.execute("UPDATE metadata SET devicename=?;",(speriod,))
        c.commit()

    def connect(self,cdb):
        if cdb is not None:
            #Connects to the given connectorDB database
            self.cdb = cdb

            if self.devicename=="":
                self.__setDeviceName(self.cdb.metaname)

    

    def addStream(self,streampath):
        #Adds the given stream to the streams that logger can cache.
        #Requires an active internet connection
        self.ensureDatabase()

        stream = None
        #Allows us to refer to streams either by the full path, or by relative path to the device
        if streampath.count("/")==0:
            stream = self.cdb[streampath]
        else:
            stream = self.cdb(streampath)
        
        if not stream.exists:
            raise errors.ServerError("The stream '%s' was not found"%(stream.metaname,))

        c = self.conn.cursor()
        c.execute("INSERT OR REPLACE INTO streams (?,?);",(stream.metaname,json.dumps(stream.schema)))
        c.commit()

        self.streams[stream.metaname] = stream.schema


    def insert(self,streamname,value):
        if streamname.count("/")==0:
            streamname = self.devicename + "/"+streamname
        if streamname.count("/")<=1 or not streamname in self.streams:
            raise errors.DataError("Stream not found '%s'"%(streamname,))

        #Validate the schema
        validate(value,self.streams[streamname])

        #Alright - all is validated. Now insert into cache
        c = self.conn.cursor()
        c.execute("INSERT INTO cache (?,?,?);",(streamname,time.time(),json.dumps(value)))
        c.commit()