import sys
import signal

from server import ServerSetup
from zookeeper import Zookeeper
from kafkaserver import Kafka


def runserver():
    s = ServerSetup(description="Connector server",bindir="./bin")
    zk = Zookeeper(s.hostname,s.port)
    kf = Kafka(zk.host,s.hostname)

    print "\n**********************************************\n"
    print "Started Successfully\nRUNNING AT",zk.host
    print "\n**********************************************\n"
    
    while (True):
        try:
            signal.pause()
        except KeyboardInterrupt:
            kf.close()
            zk.close()
            s.close()
            return 

if (__name__=="__main__"):
    runserver()