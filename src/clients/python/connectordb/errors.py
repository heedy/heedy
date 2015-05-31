"""
This file contains the errors that can be thrown
"""

#Returned when the given credentials are not accepted by the server
class AuthenticationError(Exception):
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

#Returned when the server gives an unhandled error code
class ServerError(Exception):
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)

#Returned when there is a problem connecting to the server
class ConnectionError(Exception):
    def __init__(self, value):
        self.value = value
    def __str__(self):
        return repr(self.value)