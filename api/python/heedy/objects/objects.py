from typing import Dict

from ..base import APIObject, APIList, Session
from ..kv import KV

from .. import users
from .. import apps
from ..notifications import Notifications

from . import registry


class Object(APIObject):
    props = {
        "name",
        "description",
        "icon",
        "access",
        "meta",
        "tags",
        "key",
        "owner_scope",
    }

    def __init__(self, objectData: Dict, session: Session):
        super().__init__(
            f"api/objects/{objectData['id']}",
            {"object": objectData["id"]},
            session,
            cached_data=objectData,
        )
        # The object ID
        self.id = objectData["id"]

        self._kv = KV(f"api/kv/objects/{objectData['id']}", self.session)

    @property
    def kv(self):
        return self._kv

    @kv.setter
    def kv(self, v):
        return self._kv.set(**v)

    def __getattr__(self, attr):
        try:
            return self.cached_data[attr]
        except:
            return None

    @property
    def owner(self):
        return users.User(self.cached_data["owner"], self.session)

    @property
    def app(self):
        if self.cached_data["app"] is None:
            return None
        return apps.App(self.cached_data["app"], session=self.session)

    def update(self, **kwargs):
        """
        Updates the given data::

            o.update(name="My new name",description="my new description")
        """

        meta = self.cached_data.get("meta", None)

        def updateMeta(o):
            if "result" in o and o["result"] == "ok" and "meta" in kwargs:
                if meta is None:
                    self.cached_data.pop("meta", 0)
                else:
                    # We have values of meta, so set them correctly
                    kwm = kwargs["meta"]
                    for key in kwm:
                        if kwm[key] is None:
                            meta.pop(key, 0)
                        else:
                            meta[key] = kwm[key]
                    self.cached_data["meta"] = meta
            return o

        return self.session.f(super().update(**kwargs), updateMeta)


class Objects(APIList):
    def __init__(self, constraints: Dict, session: Session):
        super().__init__("api/objects", constraints, session)

    def __getitem__(self, item):
        return super()._getitem(item, f=lambda x: registry.getObject(x, self.session))

    def __call__(self, **kwargs):
        # To query by object type, we use _type
        if "_type" in kwargs:
            kwargs["type"] = kwargs["_type"]
            del kwargs["_type"]
        return super()._call(
            f=lambda x: [registry.getObject(xx, self.session) for xx in x], **kwargs
        )

    def create(self, name, meta={}, _type="timeseries", **kwargs):
        """
        Creates a new object of the given type (timeseries by default).
        """
        return super()._create(
            f=lambda x: registry.getObject(x, self.session),
            **{"name": name, "type": _type, "meta": meta, **kwargs},
        )
