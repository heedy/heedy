import logging
logger = logging.getLogger("ZooKeeper")

import glob
import os

from serverbase import BaseServer
from kazoo.client import KazooClient


class Zookeeper(BaseServer):
    """
    Given a setup, starts the Zookeeper server with the given configuration options
    """

    def __init__(self,hostname,port=None,dbpath="./zookeeper",jardir="./bin/"):
        BaseServer.__init__(self,"zookeeper", hostname+":"+str(port),dbpath,logger,hostname,port)
        
        #Create the command line
        cmd = ["java","-Dzookeeper.log.dir="+self.folder,"-Dlog4j.configuration=file:"+os.path.join(self.folder,"log4j.properties"),
               "-cp",self.classpath(jardir),"org.apache.zookeeper.server.quorum.QuorumPeerMain",os.path.join(self.folder,"zookeeper.properties")]
        
        logproperties = self.configDefaults["log4j.properties"]
        logproperties["org.apache.zookeeper.server.DatadirCleanupManager"] = "org.apache.log4j.ConsoleAppender"
        logproperties["org.apache.zookeeper.server.DatadirCleanupManager.layout"] = "org.apache.log4j.PatternLayout"
        logproperties["log4j.appender.stdout.layout.ConversionPattern"]="[%d] %p %m (%c)%n"

        self.addConfig({"cmd": cmd,"log4j.properties": logproperties,
                        "zookeeper.properties":{
                            "dataDir": self.dbpath,
                            "clientPort": port,
                            "clientPortAddress": hostname,
                            "autopurge.snapRetainCount": 3,
                            "autopurge.purgeInterval": 1
                            }})
        self.writeConfig()
        self.runServer()

        self.connect()
        #Register!
        self.connection.registerme()

        logger.info("server running...")
            


if (__name__=="__main__"):
    import signal
    from ..setup.server import ServerSetup
    
    
    s = ServerSetup(description="Zookeeper server standalone",bindir="./bin")
    zk = Zookeeper(s.hostname,s.port)
    while (True):
        try:
            signal.pause()
        except KeyboardInterrupt:
            zk.close()
            s.close()