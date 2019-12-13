from ..base import APIObject,APIList, Session
from typing import Dict

from .. import users
from .. import apps
from ..notifications import Notifications

from . import registry

class Object(APIObject):
    props = {"name","description","icon","meta"}
    def __init__(self, objectData: Dict,  session: Session):
        super().__init__(f"api/heedy/v1/objects/{objectData['id']}", {'object': objectData['id']}, session)
        self.data = objectData



    def __getattr__(self,attr):
        return self.data[attr]

    @property
    def owner(self):
        return users.User(self.data["owner"], self.session)

    @property
    def app(self):
        if self.data["app"] is None:
            return None
        return apps.App(self.data["app"],session=self.session)


    def __str__(self):
        return str(self.data)

    def __repr__(self):
        return str(self)

class Objects(APIList):
    def __init__(self, constraints : Dict, session: Session):
        super().__init__("api/heedy/v1/objects",constraints, session)

    def __getitem__(self,item):
        return super()._getitem(item,f=lambda x : registry.getObject(x,self.session))

    def __call__(self,**kwargs):
        return super()._call(f=lambda x : [registry.getObject(xx,self.session) for xx in x],**kwargs)

    def create(self,name,meta={}, otype="stream",**kwargs):
        """
        Creates a new object of the given type (stream by default).
        """
        return super()._create(f= lambda x : registry.getObject(x,self.session) ,**{'name': name,'type': otype,'meta':meta, **kwargs})


