#This file attempts to test websocket reconnection code - the stuff which ensures that if websocket goes down, there will be a reconnect, and no subscribed values are lost.
import connectordb
import logging
import time


def test_reconnect():
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
        usr = db.getuser("python_test")
        if not usr.exists:
            usr.create("py@email","mypass")
        dev = usr["mydevice"]
        if not dev.exists:
            dev.create()

        db = connectordb.ConnectorDB("python_test/mydevice",dev.apikey,url="http://localhost:8000")

        s = db["teststream"]
        if not s.exists:
            s.create({"type": "string"})
        
        class tmpO():
            def __init__(self):
                self.gotmessages = 0
            def messagegetter(self,stream,datapoints):
                logging.info("GOT: %s",stream)
                if stream=="python_test/mydevice/teststream":
                    self.gotmessages+=1
        tmp = tmpO()
        s.subscribe(tmp.messagegetter)

        s.insert("Hello!")

        time.sleep(0.1)
        if tmp.gotmessages!=1:
            return False
    
        print "\n\n============================================"
        print "NOW RESTART THE SERVER (so that websocket dies)"
        print "type ENTER when done"
        print "============================================\n\n"
        raw_input("waiting for ENTER\n")
        print "========================================"
        print "Waiting 10 seconds for reconnect"
        print "========================================"
        time.sleep(10)

        s.insert("Hello2!")
        time.sleep(0.1)
        if tmp.gotmessages!=2:
            return False
        return True
if __name__=="__main__":
    logging.basicConfig(level=logging.DEBUG)
    status = test_reconnect()
    print "\n\n============================================"
    print "TEST STATUS:", status