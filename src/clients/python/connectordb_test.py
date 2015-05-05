import unittest
import connectordb

class TestConnectorDB(unittest.TestCase):
    def setUp(self):
        try:
            db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")
            db.user("python_test").delete()
        except:
            pass
    def test_authfail(self):
        try:
            db = connectordb.ConnectorDB("notauser","badpass",url="http://localhost:8000")
        except connectordb.AuthenticationError as e:
            return



    def test_getthis(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")

        self.assertEqual(db.name,"test/user")
        self.assertEqual(db.username,"test")
        self.assertEqual(db.devicename,"user")

    def test_adminusercrud(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")

        self.assertEqual(db.thisuser.exists,True)
        self.assertEqual(db.thisuser.admin,True)

        usr = db.user("python_test")
        self.assertFalse(usr.exists)

        usr.create("py@email","mypass")

        self.assertTrue(usr.exists)

        self.assertEqual(usr.email,"py@email")
        self.assertEqual(usr.admin,False)

        usr.email = "email@me"
        self.assertEqual(usr.email,"email@me")
        self.assertEqual(db.user("python_test").email,"email@me")
        usr.admin = True
        self.assertEqual(usr.admin,True)
        usr.admin = False
        self.assertEqual(usr.admin,False)

        self.assertRaises(connectordb.ServerError,usr.set,{"admin": "Hello"})

        self.assertEqual(len(db.users()),2)

        usr.setpassword("pass2")
        usrdb = connectordb.ConnectorDB("python_test","pass2",url="http://localhost:8000")
        self.assertEqual(usrdb.name,"python_test/user")
        usr.delete()
        self.assertFalse(db.user("python_test").exists)

        self.assertEqual(len(db.users()),1)
    def test_usercrud(self):
        db = connectordb.ConnectorDB("test","test",url="http://localhost:8000")

        usr = db.user("python_test")
        self.assertFalse(usr.exists)

        usr.create("py@email","mypass")

        db = connectordb.ConnectorDB("python_test","mypass",url="http://localhost:8000")

        self.assertEqual(len(db.users()),1,"Shouldn't see the test user")

        self.assertRaises(connectordb.AuthenticationError,db.user("hi").create,"a@b","lol")

        self.assertEqual(db.user("test").exists,False)

        self.assertRaises(connectordb.AuthenticationError,db.user("test").delete)


        usr = db.thisuser
        usr.email = "email@me"
        self.assertEqual(usr.email,"email@me")
        self.assertEqual(db.user("python_test").email,"email@me")

