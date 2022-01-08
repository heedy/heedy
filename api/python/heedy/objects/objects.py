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
        """Delete the given keys from the object metadata

        Deleting a key resets the value of that property to its default.

        .. tab:: Sync

            ::

                o.meta.delete("schema")

        .. tab:: Async

            ::

                await o.meta.delete("schema")


        Args:
            *args: The keys to delete

        Returns:
            The updated object metadata
        """
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
    """
    Object is the base class for all Heedy objects. For example, the Timeseries object type
    is a subclass of Object, and includes all of the functionality described here.

    When an object of an unrecognized type is returned from the Heedy API, it will be returned
    as the Object type.
    """

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
    """
    Each object (including Timeseries) has the above properties available as attributes.
    In synchronous sessions, they allow you to update the object's properties directly::

        o.name = "My new name"
        assert o.name == "My new name"

    The above is equivalent to::

        o.update(name="My new name")
        assert o["name"] == "My new name"

    The available properies are:
        - `name`: The name of the object, displayed as the title in the Heedy app
        - `description`: A description of the object, displayed as the subtitle in the Heedy app
        - `icon`: The icon to display in the Heedy app, either a data uri containing an image, 
          or name of icon from `Material Icons <https://fonts.google.com/icons>`_ or `Fontawesome <https://fontawesome.com/>`_.
    """

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
        """
        The user which owns this object::

            print(o.user.name) # prints your username if you own the object
        """
        return users.User(self.cached_data["owner"], self.session)

    @property
    def app(self):
        """
        The app that this object belongs to, if any::

            print(o.app.name)

        When accessing the Heedy API as an app, you will not have access to any other apps.
        """
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
