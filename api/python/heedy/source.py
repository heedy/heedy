from .base import APIObject, Session
from typing import Dict


class Object(APIObject):
    def __init__(self, session: Session, objectData: Dict):
        super().__init__(session, f"api/heedy/v1/objects/{objectData['id']}")
        self.data = objectData

    @property
    def type(self) -> str:
        """
        Returns the object type
        """
        return self.data["type"]

    def __str__(self):
        return str(self.data)

    def __repr__(self):
        return str(self)


_objectRegistry = {}


def registerObjectType(objectType: str, objectClass: Object) -> None:
    """
    registerObjectType allows external libraries to implement object types
    available through heedy plugins. All you need to do is subclass :code:`Object`,
    and register the corresponding type!
    """
    _objectRegistry[objectType] = objectClass


def getObjectObject(session: Session, objectData: Dict) -> Object:
    """
    Heedy allows multiple different object types. getObjectObject uses the
    registered object type objects to initialize the given object data to
    the correct object type. If the object is of an unregistered type, it returns
    a base :code:`Object` object.
    """
    if objectData["type"] in _objectRegistry:
        return _objectRegistry[objectData["type"]](session, objectData)
    return Object(session, objectData)
