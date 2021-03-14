from typing import Dict
from .base import APIObject, APIList, Session, getSessionType, DEFAULT_URL, q
from .kv import KV

from . import users
from . import objects
from .notifications import Notifications

from functools import partial


class App(APIObject):
    props = {"name", "description", "icon", "settings", "settings_schema"}

    def __init__(
        self, access_token: str, url: str = DEFAULT_URL, session="sync", cached_data={}
    ):
        appid = "self"
        if isinstance(session, Session):
            # Treat the session as already initialized, meaning that the access token is actually
            # the app id
            appid = access_token
            super().__init__(
                f"api/apps/{q(appid)}", {"app": appid}, session, cached_data=cached_data
            )

        else:
            # Initialize the app object as a direct API
            s = getSessionType(session, "self", url)
            s.setAccessToken(access_token)
            super().__init__("api/apps/self", {"app": appid}, s)
        # The objects belonging to the app
        self.objects = objects.Objects({"app": appid}, self.session)
        # Key-value store associated with the app
        self.kv = KV(f"api/kv/apps/{q(appid)}", self.session)

    @property
    def owner(self):
        return self.session.f(
            self.read(), lambda x: users.User(x["owner"], self.session)
        )


class Apps(APIList):
    def __init__(self, constraints: Dict, session: Session):
        super().__init__("api/apps", constraints, session)

    def __getitem__(self, item):
        return self._getitem(
            item, f=lambda x: App(x["id"], session=self.session, cached_data=x)
        )

    def __call__(self, **kwargs):
        return self._call(
            f=lambda x: [
                App(xx["id"], session=self.session, cached_data=xx) for xx in x
            ],
            **kwargs,
        )

    def create(self, name, **kwargs):
        return self._create(
            f=lambda x: App(x["id"], session=self.session, cached_data=x),
            **{"name": name, **kwargs},
        )
