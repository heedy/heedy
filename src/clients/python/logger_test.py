import unittest
import os
from connectordb.logger import ConnectorLogger

class TestLogger(unittest.TestCase):
    def setUp(self):
        if os.path.exists("test.db"):
            os.remove("test.db")
        
        l = ConnectorLogger("test.db")
        self.assertTrue(os.path.exists("test.db"))
    def test_basics(self):
        pass
        

