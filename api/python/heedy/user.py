from .base import APIObject

class User(APIObject):
    def __init__(self,session,username):
        super().__init__(session,"api/heedy/v1/users/{username}")