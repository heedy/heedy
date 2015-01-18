import logging
logger = logging.getLogger("Kafka")

import os
from serverbase import BaseServer

import time #Allows timeout for connection
from kafka import KafkaClient


class Kafka(BaseServer):
    """
    Given a folder in which a database is/should be located, it starts a kafka server rooted at that location,
    and connects to it. Once close is called, it closes the connection and kills the server.
    """
      
    def __init__(self,chost,hostname,port=None,dbpath="./kafka",jardir="./bin/"):
        BaseServer.__init__(self,"kafka",chost,dbpath,logger,hostname,port)
        self.connect()
        
        #Create base path for the kafka server (it uses zookeeper)
        self.connection.zoo.ensure_path("/kafka")

        #Create the command line
        cmd = ["java","-Xmx256M","-server","-XX:+UseCompressedOops","-XX:+UseParNewGC",
               "-XX:+UseConcMarkSweepGC","-XX:+CMSClassUnloadingEnabled","-XX:+CMSScavengeBeforeRemark",
               "-XX:+DisableExplicitGC","-Djava.awt.headless=true",
               "-Dlog4j.configuration=file:"+os.path.join(self.folder,"log4j.properties"),
                "-cp",self.classpath(jardir),"kafka.Kafka",os.path.join(self.folder,"server.properties")]
        

        logproperties = self.configDefaults["log4j.properties"]
        def addAppender(app):
            logproperties[app] = "org.apache.log4j.ConsoleAppender"
            logproperties[app+".layout"] = "org.apache.log4j.PatternLayout"
            logproperties[app+".layout.ConversionPattern"]="[%d] %p %m (%c)%n"
        addAppender("log4j.appender.kafkaAppender")
        addAppender("log4j.appender.stateChangeAppender")
        addAppender("log4j.appender.requestAppender")
        addAppender("log4j.appender.cleanerAppender")
        addAppender("log4j.appender.controllerAppender")

        self.addConfig({"cmd": cmd,"log4j.properties": logproperties,
            "server.properties": {
                    "broker.id": 0, #TODO: Set this somehow
                    "port": self.port,
                    "host.name": hostname,
                    "zookeeper.connect": chost+"/kafka", #The zookeeper host
                    "log.dirs":self.dbpath,
                    "zookeeper.connection.timeout.ms": 1000000,
                    #"num.network.threads": "2",
                    #"num.io.threads":"8",
                    "socket.send.buffer.bytes": 1048576,
                    "socket.receive.buffer.bytes": 1048576,
                    "socket.request.max.bytes": 104857600,
                    "num.partitions":"1",
                    
                    # The minimum age of a log file to be eligible for deletion
                    "log.retention.hours": 168,
                    #The maximum size of each partition is set by default to be 100MB
                    #TODO: This should be set in a more clever way to avoid filling the disk, no matter how the disk is initialized (and how many different inputs there are)
                    "log.retention.bytes": 1024*1024*100,

                    # A size-based retention policy for logs. Segments are pruned from the log as long as the remaining
                    # segments don't drop below log.retention.bytes.

                    # The maximum size of a log segment file. When this size is reached a new log segment will be created. 15MB is a good place to start.
                    #This is related to log.retention.bytes, and shoud be optimized along with it.
                    "log.segment.bytes": 1024*1024*15,

                    # The interval at which log segments are checked to see if they can be deleted according 
                    # to the retention policies
                    "log.retention.check.interval.ms": 60000,

                    # By default the log cleaner is disabled and the log retention policy will default to just delete segments after their retention expires.
                    # If log.cleaner.enable=true is set the cleaner will be enabled and individual logs can then be marked for log compaction.
                    "log.cleaner.enable":"false",

                    "auto.create.topics.enable":"true"
                }
                        })


        #Gets the configuration for the server
        self.addConfig(self.connection.config)

        self.writeConfig()
        self.runServer()
        
        #Starts the client - and gives it 2 minutes to figure out whether it is going to connect or not.
        #This is dependent on whether the database daemon is successfully starting up in the background.
        #The extremely long wait time is because some old laptops can take very long to create a database.
        #The wait time is not an issue, since we check if mongoDB crashed if we can't connect - so in effect
        #the actual wait time is at most a couple seconds if the database actually fails to start.
        t = time.time()
        while (time.time() - t < 120.0 and self.client is None):
            try:
                self.client = KafkaClient(self.host)
            except:
                time.sleep(1)
                #If the process crashed for some reason, don't continue waiting like an idiot
                if (self.server.poll() is not None):
                    logger.error("Kafka did not start correctly.")
                    self.server = None
                    break
        if (self.client==None):
            self.close()
            raise Exception("Could not connect to database")

        #The server connection is no longer necessary
        self.client.close()
        self.client = None

        #Register the client as ready
        self.connection.registerme()

        logger.info("server running...")

if (__name__=="__main__"):
    import signal
    from ..setup.server import ServerSetup
    
    
    s = ServerSetup(description="Kafka server standalone",bindir="./bin",zoorequire=True)
    db = Kafka(s.zookeeper,s.hostname,s.port)
    while (True):
        try:
            signal.pause()
        except KeyboardInterrupt:
            db.close()
            s.close()