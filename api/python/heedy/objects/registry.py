from typing import Dict
from ..base import Session

from . import objects

# This map contains the registered object types
_objectRegistry = {}


def registerObjectType(objectType: str, objectClass) -> None:
    """
    registerObjectType allows external libraries to implement object types
    available through heedy plugins. All you need to do is subclass :code:`Object`,
    and register the corresponding type!
    """
    _objectRegistry[objectType] = objectClass


def getObject(objectData: Dict, session: Session):
    """
    Heedy allows multiple different object types. getObject uses the
    registered object type to initialize the given data to
    the correct class. If the object is of an unregistered type, it returns
    a base :code:`Object` object.
    """
    if objectData["type"] in _objectRegistry:
        return _objectRegistry[objectData["type"]](objectData,session)
    return objects.Object(objectData,session)
