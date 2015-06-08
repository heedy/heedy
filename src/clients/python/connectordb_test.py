import unittest
import connectordb
from connectordb.errors import *
from jsonschema import SchemaError
import websocket
import time
import logging

websocket.enableTrace(True)

class TestConnectorDB(unittest.TestCase):
    def setUp(self):
        try:
            db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
            db.getuser("python_test").delete()
        except:
            pass
    
    def test_authfail(self):
        try:
            db = connectordb.ConnectorDB("notauser","badpass",url="http://localhost:8000")
        except AuthenticationError as e:
            return



    def test_getthis(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")

        self.assertEqual(db.name,"test/user")
        self.assertEqual(db.username,"test")
        self.assertEqual(db.devicename,"user")

    def test_adminusercrud(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")

        self.assertEqual(db.user.exists,True)
        self.assertEqual(db.user.admin,True)
        self.assertEqual(db.admin,True)

        usr = db.getuser("python_test")
        self.assertFalse(usr.exists)

        usr.create("py@email","mypass")

        self.assertTrue(usr.exists)

        self.assertEqual(usr.email,"py@email")
        self.assertEqual(usr.admin,False)

        usr.email = "email@me"
        self.assertEqual(usr.email,"email@me")
        self.assertEqual(db.getuser("python_test").email,"email@me")
        usr.admin = True
        self.assertEqual(usr.admin,True)
        usr.admin = False
        self.assertEqual(usr.admin,False)

        self.assertRaises(ServerError,usr.set,{"admin": "Hello"})

        self.assertEqual(len(db.users()),2)

        usr.setpassword("pass2")
        usrdb = connectordb.ConnectorDB("python_test","pass2",url="http://localhost:8000")
        self.assertEqual(usrdb.name,"python_test/user")
        usr.delete()
        self.assertFalse(db.getuser("python_test").exists)

        self.assertEqual(len(db.users()),1)
    def test_usercrud(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")

        usr = db.getuser("python_test")
        self.assertFalse(usr.exists)

        usr.create("py@email","mypass")

        db = connectordb.ConnectorDB("python_test","mypass",url="http://localhost:8000")

        self.assertEqual(len(db.users()),1,"Shouldn't see the test user")

        self.assertRaises(AuthenticationError,db.getuser("hi").create,"a@b","lol")

        self.assertEqual(db.getuser("test").exists,False)

        self.assertRaises(AuthenticationError,db.getuser("test").delete)


        usr = db.user
        usr.email = "email@me"
        self.assertEqual(usr.email,"email@me")
        self.assertEqual(db.getuser("python_test").email,"email@me")

    def test_devicecrud(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
        usr = db.getuser("python_test")
        usr.create("py@email","mypass")

        db = connectordb.ConnectorDB("python_test","mypass",url="http://localhost:8000")

        self.assertTrue(db.exists)
        self.assertEqual(1,len(db.user.devices()))

        self.assertFalse(db.user["mydevice"].exists)
        db.user["mydevice"].create()

        self.assertTrue(db.user["mydevice"].exists)

        self.assertEqual(2,len(db.user.devices()))

        db = connectordb.ConnectorDB("python_test/mydevice",db.user["mydevice"].apikey,url="http://localhost:8000")

        self.assertEqual(1,len(db.user.devices()))

        db.nickname = "testnick"
        self.assertEqual(db.nickname,"testnick")
        self.assertEqual(db.user.email,"py@email")
        self.assertRaises(AuthenticationError,db.delete)

        newkey = db.resetKey()
        self.assertRaises(AuthenticationError,db.refresh)


        db = connectordb.ConnectorDB("python_test/mydevice",newkey,url="http://localhost:8000")
        self.assertTrue(db.exists)

    def test_streamcrud(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
        usr = db.getuser("python_test")
        usr.create("py@email","mypass")
        dev = usr["mydevice"]
        dev.create()

        self.assertTrue(dev.exists)
        db = connectordb.ConnectorDB("python_test/mydevice",dev.apikey,url="http://localhost:8000")


        self.assertTrue(db.exists)

        s = db["mystream"]

        self.assertRaises(SchemaError,s.create,{"type":"blah blah"})

        s.create({"type":"string"})
        self.assertTrue(s.exists)
        s.delete()
        self.assertFalse(s.exists)

        db["mystream"].create({"type":"string"})
        self.assertTrue(s.exists)
        s.name = "differentstream"
        self.assertFalse(s.exists)

        self.assertTrue(db["differentstream"].exists)
        print db["differentstream"].metadata
        self.assertEqual(len(db.streams()),1)
        self.assertEqual(db["differentstream"].schema["type"],"string")
        self.assertEqual(db["differentstream"].name,"differentstream")
        db["differentstream"].delete()
        self.assertFalse(db["differentstream"].exists)
        self.assertEqual(len(db.streams()),0)

    def test_streamio(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
        usr = db.getuser("python_test")
        usr.create("py@email","mypass")
        dev = usr["mydevice"]
        dev.create()

        self.assertTrue(dev.exists)
        db = connectordb.ConnectorDB("python_test/mydevice",dev.apikey,url="http://localhost:8000")

        s = db["teststream"]

        self.assertFalse(s.exists)

        s.create({"type": "string"})
        self.assertTrue(s.exists)

        self.assertEqual(0,len(s))

        s.insert("Hello World!")

        self.assertEqual(1,len(s))

        self.assertEqual("Hello World!",s[0]["d"])
        self.assertEqual("Hello World!",s(0)[0]["d"])

    
    def test_subscribe(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
        usr = db.getuser("python_test")
        usr.create("py@email","mypass")
        dev = usr["mydevice"]
        dev.create()

        self.assertTrue(dev.exists)
        db = connectordb.ConnectorDB("python_test/mydevice",dev.apikey,url="http://localhost:8000")

        s = db["teststream"]
        s.create({"type": "string"})
        
        class tmpO():
            def __init__(self):
                self.gotmessage = False
            def messagegetter(self,stream,datapoints):
                logging.info("GOT: %s",stream)
                if stream=="python_test/mydevice/teststream":
                    logging.info("SETTING TRUE")
                    self.gotmessage=True
        tmp = tmpO()
        s.subscribe(tmp.messagegetter)

        s.insert("Hello!")

        time.sleep(0.1)
        logging.info("Checking Truth")
        self.assertTrue(tmp.gotmessage)
    
        tmp.gotmessage=False

        s.unsubscribe()

        s.insert("Hello Again!")

        time.sleep(0.1)
        self.assertFalse(tmp.gotmessage)

    def test_wsinsert(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
        usr = db.getuser("python_test")
        usr.create("py@email","mypass")
        dev = usr["mydevice"]
        dev.create()

        self.assertTrue(dev.exists)
        db = connectordb.ConnectorDB("python_test/mydevice",dev.apikey,url="http://localhost:8000")

        s = db["teststream"]

        self.assertFalse(s.exists)

        s.create({"type": "string"})
        self.assertTrue(s.exists)

        self.assertEqual(0,len(s))

        #Use websocket for inserts
        db.wsinsert = True

        s.insert("Hello World!")

        self.assertEqual(1,len(s))

        self.assertEqual("Hello World!",s[0]["d"])
        self.assertEqual("Hello World!",s(0)[0]["d"])