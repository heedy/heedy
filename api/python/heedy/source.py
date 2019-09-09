from .base import APIObject, Session
from typing import Dict



class Source(APIObject):
    def __init__(self,session : Session, sourceData : Dict):
        super().__init__(session,f"api/heedy/v1/sources/{sourceData['id']}")
        self.data = sourceData

    @property
    def type(self) -> str:
        """
        Returns the source type
        """
        return self.data["type"]
    def __str__(self):
        return str(self.data)
    def __repr__(self):
        return str(self)



_sourceRegistry = {}

def registerSourceType(sourceType : str, sourceClass : Source) -> None:
    """
    registerSourceType allows external libraries to implement source types
    available through heedy plugins. All you need to do is subclass :code:`Source`,
    and register the corresponding type!
    """
    _sourceRegistry[sourceType] = sourceClass

def getSourceObject(session : Session, sourceData: Dict) -> Source:
    """
    Heedy allows multiple different source types. getSourceObject uses the
    registered source type objects to initialize the given source data to
    the correct object type. If the source is of an unregistered type, it returns
    a base :code:`Source` object.
    """
    if sourceData["type"] in _sourceRegistry:
        return _sourceRegistry[sourceData["type"]](session,sourceData)
    return Source(session,sourceData)
