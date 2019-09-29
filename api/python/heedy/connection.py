from .base import APIObject, getSessionType, DEFAULT_URL
from .source import getSourceObject

from functools import partial


class Connection(APIObject):
    def __init__(self, access_token: str, url: str = DEFAULT_URL, session: str = "sync"):
        # Initialize the connection object
        s = getSessionType(session)
        s.setAccessToken(access_token)
        super().__init__(s, "api/heedy/v1/connections/self")

    def createSource(self, **kwargs):
        return self.session.post("api/heedy/v1/sources", data=kwargs, f=lambda x: getSourceObject(self.session, x))

    def listSources(self, **kwargs):
        return self.session.get("api/heedy/v1/sources",
                                params=kwargs, f=lambda x: list(map(partial(getSourceObject, self.session), x)))

    def notify(self, n, **kwargs):
        return self.session.post("/api/heedy/v1/notifications", n, params=kwargs)
