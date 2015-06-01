import unittest
import os
import time
import connectordb
from connectordb.logger import ConnectorLogger

class TestLogger(unittest.TestCase):
    def setUp(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
        try:
            db.getuser("python_test").delete()
        except:
            pass
        usr = db.getuser("python_test")
        usr.create("py@email","mypass")
        usr["mydevice"].create()
        self.apikey = usr["mydevice"].apikey

        if os.path.exists("test.db"):
            os.remove("test.db")
        
        self.l = ConnectorLogger("test.db")
        self.assertTrue(os.path.exists("test.db"))

    def test_inserting(self):
        l = self.l
        db = connectordb.ConnectorDB("python_test/mydevice",self.apikey,url="http://localhost:8000")
        l.connect(db)
        self.assertEqual(l.devicename,"python_test/mydevice")
        s = db["mystream"]
        s.create({"type":"string"})

        l.addStream("mystream")
        
        self.assertTrue("python_test/mydevice/mystream" in l.streams)

        self.assertEqual(0,len(l))

        l.insert("mystream","Hello World!")

        self.assertEqual(1,len(l))

        l = None
        self.l = None

        #Now reload from file, and make sure everything was saved
        l = ConnectorLogger("test.db")
        self.assertEqual(1,len(l))
        self.assertEqual(l.devicename,"python_test/mydevice")
        self.assertTrue("python_test/mydevice/mystream" in l.streams)

        l.connect(db)
        self.assertEqual(0,len(s))
        l.sync()
        self.assertEqual(1,len(s))
        self.assertEqual(0,len(l))
        l.sync()
        self.assertEqual(1,len(s))
        self.assertEqual(0,len(l))

        self.assertGreater(l.lastsync,time.time()-1.)

    def test_bgsync(self):
        l = self.l
        db = connectordb.ConnectorDB("python_test/mydevice",self.apikey,url="http://localhost:8000")
        l.connect(db)
        s = db["mystream"]
        s.create({"type":"string"})

        l.syncperiod=1
        l.addStream("mystream")

        self.assertEqual(0,len(s))
        self.assertEqual(0,len(l))

        l.run()
        l.insert("mystream","hi")
        l.insert("mystream","hello")
        self.assertEqual(0,len(s))
        self.assertEqual(2,len(l))
        time.sleep(1.3)
        self.assertEqual(2,len(s))
        self.assertEqual(0,len(l))
        l.stop()