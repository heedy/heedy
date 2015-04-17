import unittest
import connectordb

class TestConnectorDB(unittest.TestCase):
    def test_authfail(self):
        try:
            db = connectordb.ConnectorDB("notauser",url="http://localhost:8000")
        except connectordb.AuthenticationError as e:
            return



        