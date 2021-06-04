from typing import Dict
import pprint

from ..base import APIObject, APIList, Session, q
from ..kv import KV

from .. import users
from .. import apps
from ..notifications import Notifications

from . import registry


class ObjectMeta:
    """ObjectMeta is a wrapper class that makes metadata access more pythonic, allowing simple updates
    such as::

        o.meta.schema = {"type":"number"}

    """

    def __init__(self, obj):
        super().__setattr__("_object", obj)

    @property
    def cached_data(self):
        return self._object.cached_data["meta"]

    def update(self, **kwargs):
        """Update the given elements of object metadata"""
        return self._object.update({"meta": kwargs})

    def delete(self, *args):
        toDelete = {}
        for a in args:
            toDelete[a] = None
        return self._object.update(meta=toDelete)

    def __getattr__(self, attr):
        return self.cached_data[attr]

    def __getitem__(self, i):
        # Gets the item from the cache - assumes that the data is in the cache. If not, need to call .read() first
        return self.cached_data[i]

    def __setitem__(self, name, value):
        return self._object.update(meta={name: value})

    def __delitem__(self, name):
        return self.delete(name)

    def __setattr__(self, name, value):
        return self.__setitem__(name, value)

    def __delattr__(self, name):
        return self.__delitem__(name)

    def __iter__(self):
        return iter(self.cached_data)

    def __contains__(self, item):
        return item in self.cached_data

    def __len__(self):
        return len(self.cached_data)

    def __str__(self):
        return pprint.pformat(self.cached_data)

    def __repr__(self):
        return str(self)


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
            f"api/objects/{q(objectData['id'])}",
            {"object": objectData["id"]},
            session,
            cached_data=objectData,
        )
        # The object ID
        self.id = objectData["id"]

        self._kv = KV(f"api/kv/objects/{q(objectData['id'])}", self.session)

    @property
    def kv(self):
        return self._kv

    @kv.setter
    def kv(self, v):
        return self._kv.set(**v)

    @property
    def meta(self):
        return ObjectMeta(self)

    @meta.setter
    def meta(self, v):
        return self.update(meta=v)

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
        return super()._call(
            f=lambda x: [registry.getObject(xx, self.session) for xx in x], **kwargs
        )

    def create(self, name, meta={}, type="timeseries", **kwargs):
        """
        Creates a new object of the given type (timeseries by default).
        """
        return super()._create(
            f=lambda x: registry.getObject(x, self.session),
            **{"name": name, "type": type, "meta": meta, **kwargs},
        )
