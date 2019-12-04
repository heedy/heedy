from .base import APIObject, getSessionType, DEFAULT_URL
from .object import getObject

from functools import partial


class App(APIObject):
    def __init__(self, access_token: str, url: str = DEFAULT_URL, session: str = "sync"):
        # Initialize the app object
        s = getSessionType(session)
        s.setAccessToken(access_token)
        super().__init__(s, "api/heedy/v1/apps/self")

    def createObject(self, **kwargs):
        return self.session.post("api/heedy/v1/objects", data=kwargs, f=lambda x: getObject(self.session, x))

    def listObjects(self, **kwargs):
        return self.session.get("api/heedy/v1/objects",
                                params=kwargs, f=lambda x: list(map(partial(getObject, self.session), x)))

    def notify(self, key, title, **kwargs):
        n = {
            "key": key,
            "title": title,
            **kwargs
        }
        if "_global" in n:
            n["global"] = n["_global"]
            del n["_global"]

        return self.session.post("/api/heedy/v1/notifications", n)

    def delete_notification(self, key=None, **kwargs):
        if key is not None:
            kwargs["key"] = key
        return self.session.delete("/api/heedy/v1/notifications", params=kwargs)
