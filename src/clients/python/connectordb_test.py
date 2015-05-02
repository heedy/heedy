import unittest
import connectordb

class TestConnectorDB(unittest.TestCase):
    def test_authfail(self):
        try:
            db = connectordb.ConnectorDB("notauser",url="http://localhost:8000")
        except connectordb.AuthenticationError as e:
            return



    def test_getthis(self):
        db = connectordb.ConnectorDB("test",username="test",url="http://localhost:8000")

        self.assertEqual(db.name,"test/user")
        self.assertEqual(db.username,"test")
        self.assertEqual(db.devicename,"user")

