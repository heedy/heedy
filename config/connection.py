
import logging

logger = logging.getLogger("connector.connection")

from kazoo.client import KazooClient, KazooState

from bson import BSON

import socket   #The hostname is part of socket
import os       #Can get PID thru os

class Connection(object):
    def __init__(self,zoohosts,cname="client",hostname="localhost",rootpath="/connector"):
        logger.info("connecting as '%s'...",cname)

        self.name = hostname + ":" + str(os.getpid())
        self.cname = cname
        self.rootpath = rootpath

        #Connect to the zookeeper instance
        self.zoo = KazooClient(hosts=zoohosts)
        self.zoo.add_listener(self.state_listener)
        self.zoo.start()

        #self.registerme()


    def registerme(self):
        #Notifies the zookeeper that this server is connected
        basepath = self.rootpath+"/"+self.cname

        logger.info("registering %s at %s",self.name,basepath)

        #Creates the entire pathway to db list
        self.zoo.ensure_path(basepath)
        #Create ephemeral node which represents connection status of this node
        self.zoo.create(basepath+"/"+self.name,ephemeral=True)
        

    def state_listener(self,state):
        """
        if (state == KazooState.CONNECTED):
            logger.info("connected")
        elif (state == KazooState.SUSPENDED):
            logger.error("suspended connecton")
        else:
            logger.error("connection lost")
        """
        pass

    def close(self):
        if (self.zoo is not None):
            logger.info("closing connection")
            self.zoo.stop()
        self.zoo = None
    def __del__(self):
        if (self.zoo is not None):
            self.close()

    def getconfiguration(self,name=None):
        if (name is None):
            name = self.cname

        logger.info("Getting config for '%s'",name)

        #Load the general configuration for objects of the given type
        # - need to handle error path not exist
        try:
            data,stat = self.zoo.get(self.rootpath+"/"+name)
        except: #An exceptino happens if the path does not exist
            return None
        if (data!=""):
            return BSON.decode(data)
        else:
            return None

    def setconfiguration(self,config,name=None):
        if (name is None):
            name = self.cname

        logger.info("Setting config for '%s'",name)

        #Set the general configuration for objects of the given type
        self.zoo.set(self.rootpath+"/"+name,BSON.encode(config))
        
    config = property(getconfiguration,setconfiguration)
