from typing import Dict
from ..base import Session

from . import objects

# This map contains the registered object types
_objectRegistry = {}


def registerObjectType(objectType: str, objectClass):
    """
    If implementing a plugin which creates a new object type in Heedy, it might be useful to add support for the object type in the heedy python client.
    This is done by creating a subclass of :code:`heedy.objects.Object`, and then registering it::

        from heedy.objects import Object,registerObjectType

        class MyType(Object):
            def myfunction(self):
                # Returns the result of a REST API call for the object
                return self.session.get(self.uri + "/mytype/my_rest_endpoint")

        registerObjectType("mytype",MyType)

    Then, when reading objects, the object type is automatically detected and the correct class is used:

    .. tab:: Sync

        ::

            mtobjs = app.objects(type="mytype")

            for mtobj in mtobjs:
                print(mtobj.myfunction())

    .. tab:: Async

        ::

            mtobjs = await app.objects(type="mytype")

            for mtobj in mtobjs:
                print(await mtobj.myfunction())

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
        return _objectRegistry[objectData["type"]](objectData, session)
    return objects.Object(objectData, session)
