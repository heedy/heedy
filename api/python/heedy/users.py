from typing import Dict
from .base import APIObject, APIList, Session, q
from .kv import KV

from . import apps
from . import objects
from .notifications import Notifications


class User(APIObject):
    props = {"name", "username", "description", "icon"}

    def __init__(self, username: str, session: Session, cached_data={}):
        super().__init__(
            f"api/users/{q(username)}",
            {"user": username},
            session,
            cached_data=cached_data,
        )
        self._username = username

        # Apps represents a list of the user's active apps. Apps can be accessed by ID::
        #   myapp = await u.apps["appid"]
        # They can also by queried, which will return a list::
        #   myapp = await u.apps(plugin="myplugin")
        # Finally, they can be created::
        #   myapp = await u.apps.create()
        self.apps = apps.Apps({"owner": username}, self.session)
        self.objects = objects.Objects({"owner": username}, self.session)
        self._kv = KV(f"api/kv/users/{q(username)}", self.session)

    @property
    def kv(self):
        return self._kv

    @property
    def username(self):
        return self._username

    @kv.setter
    def kv(self, v):
        return self._kv.set(**v)


class Users(APIList):
    def __init__(self, constraints: Dict, session: Session):
        super().__init__("api/users", constraints, session)

    def __getitem__(self, item):
        return self._getitem(
            item, f=lambda x: User(x["username"], session=self.session, cached_data=x)
        )

    def __call__(self, **kwargs):
        return self._call(
            f=lambda x: [
                User(xx["id"], session=self.session, cached_data=xx) for xx in x
            ],
            **kwargs,
        )

    def create(self, username, password, **kwargs):
        return self._create(
            f=lambda x: User(x["username"], session=self.session, cached_data=x),
            **{"username": username, "password": password, **kwargs},
        )
