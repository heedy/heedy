import unittest
import os
import time
import connectordb
from connectordb.logger import ConnectorLogger

class TestLogger(unittest.TestCase):
    def setUp(self):
        self.url = "http://localhost:8000"
        db = connectordb.ConnectorDB("test","test",url=self.url)
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
        

    def test_inserting(self):
        db = connectordb.ConnectorDB("test","test",url=self.url)
        s = db("python_test/mydevice/mystream")
        self.assertFalse(s.exists)

        def test_create(obj):
            obj.setlogin("python_test/mydevice",self.apikey,self.url)

            obj.addStream("mystream",{"type":"string"})

            haderror=False
            try:
                obj.addStream("badstream")
            except:
                haderror=True
            self.assertTrue(haderror)


        l = ConnectorLogger("test.db",on_create=test_create)

        self.assertTrue(s.exists)

        self.assertEqual("python_test/mydevice",l.name)
        self.assertEqual(self.apikey,l.apikey)
        self.assertEqual(self.url,l.url)

        self.assertEqual(0,len(l))

        self.assertTrue("python_test/mydevice/mystream" in l)
        self.assertFalse("python_test/mydevice/badstream" in l)

        l.insert("mystream","Hello World!")

        self.assertEqual(1,len(l))

        l.data = "Hello World!"

        self.assertEqual("Hello World!",l.data)

        l = None
        self.l = None

        #Now reload from file, and make sure everything was saved
        l = ConnectorLogger("test.db")
        self.assertEqual(1,len(l))
        self.assertEqual(l.name,"python_test/mydevice")
        self.assertTrue("python_test/mydevice/mystream" in l)

        self.assertEqual(0,len(s))
        l.sync()
        self.assertEqual(1,len(s))
        self.assertEqual(0,len(l))
        l.sync()
        self.assertEqual(1,len(s))
        self.assertEqual(0,len(l))

        self.assertGreater(l.lastsync,time.time()-1.)

        self.assertEqual("Hello World!",l.data)

    def test_bgsync(self):
        db = connectordb.ConnectorDB("test","test",url=self.url)
        s = db("python_test/mydevice/mystream")
        self.assertFalse(s.exists)

        l = ConnectorLogger("test.db")
        l.setlogin("python_test/mydevice",self.apikey,self.url)
        l.addStream("mystream",{"type":"string"})

        l.syncperiod=1

        self.assertEqual(0,len(s))
        self.assertEqual(0,len(l))

        l.start()
        l.insert("mystream","hi")
        l.insert("mystream","hello")
        self.assertEqual(0,len(s))
        self.assertEqual(2,len(l))
        time.sleep(1.3)
        self.assertEqual(2,len(s))
        self.assertEqual(0,len(l))
        l.stop()