import logging
import os
import glob
import signal
import shutil
import threading
from subprocess32 import Popen,PIPE,STDOUT,TimeoutExpired

from connection import Connection
from server import get_open_port

class BaseServer(object):
    def __init__(self,name,zoohost,folder,logger,hostname,port=None,minspace=0.):
        logger.info("Starting server...")
        self.logger=logger
        self.port = port
        #If it doesn't matter what port to use, just use a random one
        if (self.port is None):
            self.port = get_open_port()
        self.name = name
        self.folder = folder
        self.dbpath = os.path.join(self.folder,"db")
        self.zoohost = zoohost
        self.host = hostname+":"+str(self.port)
        
        

        self.configDefaults = {
        "log4j.properties": {
            "log4j.rootLogger": "INFO, stdout",
            "log4j.appender.stdout": "org.apache.log4j.ConsoleAppender",
            "log4j.appender.stdout.layout": "org.apache.log4j.PatternLayout",
            "log4j.appender.stdout.layout.ConversionPattern":"[%d] %p %m (%c)%n"
            }
        }
        self.config = {"cmd": []}
        
        #These three are the three necessary parts to the server. The self.connection is handled
        #   in baseServer, but the rest are not.
        self.connection = None
        self.client = None
        self.server = None
        
        #Make sure the folder exists
        if (self.ensureFolder(self.folder) and minspace > 0.):
            try:
                self.checkDiskSpace(minspace)
            except:
                self.deleteFolder(self.folder)
        self.ensureFolder(self.dbpath)

    #Runs the given command
    def runServer(self,cmd=None):
        if (cmd is None):
            cmd = self.config["cmd"]
        self.logger.info("Running command: %s",str(cmd))
        self.server = Popen(cmd,stdout=PIPE,stderr=STDOUT,start_new_session=True)  #Don't forward signals
        self.logt = threading.Thread(target=self.logserver)
        self.logt.daemon = False
        self.logt.start()

    def logserver(self):
        while self.server is not None:
            line = self.server.stdout.readline()
            if not line:
                break
            self.logger.info("-- "+ line.strip())

    def connect(self):
        self.connection = Connection(self.zoohost,self.name,self.host)

    def ensureFolder(self,folder):
        self.logger.info("Checking '%s'",folder)

        if not (os.path.isdir(folder)):
            if (os.path.exists(folder)):
                self.logger.error("'%s' is not a directory")
                raise Exception(folder+" is not a directory!")
            else:
                self.logger.warning("Creating path '%s'",folder)
                os.mkdir(folder,0700)
                return True
        return False

    def deleteFolder(self,folder):
        shutil.rmtree(folder)

    def checkDiskSpace(self,minspace=0.0):
        disk = os.statvfs(self.folder)
        disk = float(disk.f_frsize*disk.f_bavail)/1024/1024

        if (disk < minspace):
            self.logger.error("Not enough free space to run database! (%.2fMB/%.2fMB)",disk,minspace)
        return disk
    
    def classpath(self,jardir):
        #Returns the list of jar files ready to be input into java
        classpath = ""
        for g in glob.glob(os.path.join(jardir,"*.jar")):
            classpath = classpath + ":"+g
        return classpath[1:]   #get rid of starting :

    def close(self,waitTime=10.):
        self.logger.warn("Shutting down server...")
        if (self.connection is not None):
            self.connection.close()
        if (self.client is not None):
            self.client.close()
        if (self.server is not None):
            self.server.send_signal(signal.SIGINT)
            try:
                self.server.wait(waitTime)
            except TimeoutExpired:
                self.logger.warn("Expired close timeout - killing process")
                self.server.kill()


        self.client = None
        self.connection = None
        self.server = None

        self.logger.info("server shutdown complete")

    def __del__(self):
        if (self.server is not None or self.connection is not None or self.client is not None):
            self.close()

    #Configuration file details
    def addConfig(self,cfg):
        if (cfg is None): 
            return    #Make sure that the config is not None
        for file in cfg:
            if (file=="cmd"):
                self.logger.info("Updating server cmd")
                self.config["cmd"] = self.config["cmd"] + cfg[file]
            else:
                self.logger.info("Updating configuration file '%s'",file)
                if (isinstance(cfg[file],dict)):
                    if (file not in self.config):
                        self.config[file]={}
                    for property in cfg[file]:
                        self.config[file][property] = cfg[file][property]
                else:
                    self.config[file] = str(cfg[file])

    #Write generic java configuration file (useful for zookeeper and kafka)
    def writeConfig(self):
        self.logger.info("Writing configuration...")
        for file in self.config:
            if (file!="cmd"):   #The cmd refers to command line arguments
                self.logger.info("write config: '%s'",file)
                f = open(os.path.join(self.folder,file),"w")
                if (isinstance(self.config[file],dict)):
                    for property in self.config[file]:
                        f.write(str(property)+"="+str(self.config[file][property])+"\n")
                else:
                    f.write(self.config[file])
                f.close()