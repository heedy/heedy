from typing import Dict
import pprint

from ..base import APIObject, APIList, Session, q
from ..kv import KV

from .. import users
from .. import apps
from ..notifications import Notifications

from . import registry


class ObjectMeta:
    """
    Heedy's objects have a metadata field which stores object type-specific information.
    This :code:`meta` property of an object is a key-value dictionary, and can be edited by altering
    the meta field in sync sessions, or by calling the :code:`update` method.

    .. tab:: Sync

        ::

            # A timeseries type object has a schema key in its metadata. Setting "schema" here
            # does not alter any other elements of :code:`meta`, but only the "schema" key.
            o.meta = {"schema": {"type": "string"}}

    .. tab:: Async

        ::

            # A timeseries type object has a schema key in its metadata. Setting "schema" here
            # does not alter any other elements of :code:`meta`, but only the "schema" key.
            await o.update(meta={"schema": {"type": "string"}})

    The :code:`meta` property behaves like a dictionary, but has features that help in usage. For example,
    the above code can be written as:

    .. tab:: Sync

        ::

            o.meta.schema = {"type": "string"}

    .. tab:: Async

        ::

            await o.meta.update(schema={"type": "string"})


    The :code:`meta` property also has several properties that offer syntactic sugar in synchronous code:

    .. tab:: Sync

        ::

            del o.meta.schema # Resets the schema to its default value
            del o.meta["schema"] # Same as above
            len(o.meta) # Returns the number of keys currently cached in the object metadata
            "schema" in o.meta # Returns True if the schema key is in the cached meta field


    .. tab:: Async

        ::

            await o.meta.delete("schema") # Resets the schema to its default value

            len(o.meta) # Returns the number of keys currently cached in the object metadata
            "schema" in o.meta # Returns True if the schema key is in the cached meta field


    Finally, the semantics of :code:`o.meta.schema` and :code:`o.meta["schema"]` are the same as for standard objects,
    meaning that :code:`o.meta["schema"]` does not query the server for the schema, but instead returns the cached values,
    while :code:`o.meta.schema` will always query the server, and needs to be awaited in async sessions.
    """

    def __init__(self, obj):
        super().__setattr__("_object", obj)

    @property
    def cached_data(self):
        return self._object.cached_data["meta"]

    def update(self, **kwargs):
        """Sets the given keys in the object's type metadata.

        .. tab:: Sync

            ::

                o.meta.update(schema={"type": "string"})

        .. tab:: Async

            ::

                await o.meta.update(schema={"type": "string"})

        Args:
            **kwargs: The keys to set and their values
        Returns:
            The updated object metadata (as a dictionary)
        Raises:
            HeedyException: If writing fails (usually due to insufficient permissions)

        """
        return self._object.update(meta=kwargs)

    def delete(self, *args: str):
        """Delete the given keys from the object metadata.

        Deleting a key resets the value of that property to its default.
        Removes the key from metadata if it is optional.

        .. tab:: Sync

            ::

                o.meta.delete("schema")
                assert o.meta["schema"] == {}

        .. tab:: Async

            ::

                await o.meta.delete("schema")
                assert o.meta["schema"] == {}


        Args:
            *args: The keys to delete

        Returns:
            The updated object metadata
        Raises:
            HeedyException: If writing fails (usually due to insufficient permissions)
        """
        toDelete = {}
        for a in args:
            toDelete[a] = None
        return self._object.update(meta=toDelete)

    def __call__(self):
        """
        Queries the server for the object's metadata, and returns it as a dictionary.
        This does not need to be called if the object was already read, since an object
        read automatically caches its metadata.

        .. tab:: Sync

            ::

                metadict = o.meta()

        .. tab:: Async

            ::

                metadict = await o.meta()

        Usually, you will not need to call this method, since ObjectMeta behaves like a dictionary
        once the object is read.

        Returns:
            The object's metadata as a dictionary
        """
        return self._object.session.f(self._object.read(), lambda o: o["meta"])

    def __getattr__(self, attr):
        return self._object.session.f(self._object.read(), lambda o: o["meta"][attr])

    def __getitem__(self, key: str):
        return self.cached_data[key]

    def __setitem__(self, name: str, value):
        return self._object.update(meta={name: value})

    def __delitem__(self, name: str):
        return self.delete(name)

    def __setattr__(self, name: str, value):
        return self.__setitem__(name, value)

    def __delattr__(self, name: str):
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

    def __eq__(self, o: object) -> bool:
        if isinstance(o, ObjectMeta):
            return self._object == o._object
        else:
            return self.cached_data == o


class Object(APIObject):
    """
    Object is the base class for all Heedy objects. For example, the Timeseries object type
    is a subclass of Object, and therefore includes all of the functionality described here.

    When an object of an unrecognized type is returned from the Heedy API, the Python client
    will return it as the Object type.
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

    def __init__(self, cached_data: Dict, session: Session):
        super().__init__(
            f"api/objects/{q(cached_data['id'])}",
            {"object": cached_data["id"]},
            session,
            cached_data=cached_data,
        )

        self._kv = KV(f"api/kv/objects/{q(cached_data['id'])}", self.session)

    @property
    def id(self):
        """
        The object's unique ID. This is directly available, so does not need
        to be awaited in async sessions::

            print(myobj.id)
        """
        return self.cached_data["id"]

    @property
    def kv(self):
        """
        The key-value store associated with this object. For details of usage, see :ref:`python_kv`.

        Returns:
            A :class:`heedy.kv.KV` object for the element.
        """
        return self._kv

    @kv.setter
    def kv(self, v):
        return self._kv.set(**v)

    @property
    def meta(self):
        """
        The object type's metadata. For details of usage, see :ref:`python_objectmeta`.
        This does not need to be awaited in async sessions. Instead, to read the data if it was
        not yet cached, use :code:`o.meta()` or :code:`await o.meta()` depending on session type.


        Returns:
            An :class:`heedy.objects.objects.ObjectMeta` object for this element. See :ref:`python_objectmeta`.
        """
        return ObjectMeta(self)

    @meta.setter
    def meta(self, v):
        return self.update(meta=v)

    def __getitem__(self, i):
        if i == "meta":
            return self.meta
        v = super().__getitem__(i)
        if i == "owner":
            return users.User(v, self.session)
        if i == "app" and v is not None:
            return apps.App(v, session=self.session)
        
        return v

    @property
    def owner(self):
        """
        The user which owns this object:

        .. tab:: Sync

            ::

                print(o.owner.username) # Queries the server for the owner, and prints username
                print(o["owner"].username) # Uses cached query data to get the owner


        .. tab:: Async

            ::

                print((await o.owner).username) # Queries the server for the owner, and prints username
                print(o["owner"].username) # Uses cached query data to get the owner



        Returns:
            The :class:`~heedy.User` object of the user which owns this object.
            The returned user does not have any cached data other than its username.
        """
        return self.session.f(
            self.read(), lambda x: users.User(x["owner"], self.session)
        )

    @property
    def app(self):
        """
        The app that manages this object, if any:

        .. tab:: Sync

            ::

                if obj.app is not None:  # Query the server for the object's app
                    print(obj["app"].id) # Use cached value from previous query


        .. tab:: Async

            ::

                if await obj.app is not None: # Query the server for the object's app
                    print(obj["app"].id) # Use cached value from previous query

        When accessing the Heedy API as an app, you will not be able to see or access any other apps.

        Returns:
            The app that this object belongs to, or None if it does not belong to an app.
            The returned :class:`~heedy.App` object does not have any cached data other than its id, so you will need to call
            :code:`read()` on it if accessing more than its id.
        """
        return self.session.f(
            self.read(),
            lambda x: apps.App(x["app"], session=self.session)
            if x["app"] is not None
            else None,
        )

    def update(self, **kwargs):

        meta = self.cached_data.get("meta", {})

        def updateMeta(o):
            if "result" in o and o["result"] == "ok" and "meta" in kwargs:
                if kwargs["meta"] is None:
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
    """
    Objects is a class implementing a list of objects. It is accessed as a property
    of users/apps/plugin to get the objects belonging to that user/app, or to query
    objects in all of heedy using plugin.

    .. tab:: Sync

        ::

            myuser.objects() # objects belonging to myuser
            myapp.objects() # objects managed by myapp
            plugin.objects() # all objects in heedy

    .. tab:: Async

        ::

            await myuser.objects() # objects belonging to myuser
            await myapp.objects() # objects managed by myapp
            await plugin.objects() # all objects in heedy


    """

    def __init__(self, constraints: Dict, session: Session):
        super().__init__("api/objects", constraints, session)

    def __getitem__(self, objectId: str):
        """Gets an object by its ID. Each object in heedy has a unique string ID,
        which can then be used to access the object.
        The ID can be seen in the URL of the object's page in the frontend.

        .. tab:: Sync

            ::

                obj = p.objects["d233rk43o6kkle43kl"]

        .. tab:: Async

            ::

                obj = await p.objects["d233rk43o6kkle43kl"]

        Returns:
            The :class:`~heedy.Object` with the given ID (or promise for the object)
        Throws:
            HeedyException: If the object does not exist
        """
        return super()._getitem(
            objectId, f=lambda x: registry.getObject(x, self.session)
        )

    def __call__(self, **kwargs):
        """Gets the objects matching the given constraints.
        If used as a property of a user of an app,
        it will return only the objects belonging to that user/app (the user/app constraint is automatically added).

        .. tab:: Sync

            ::

                obj = p.objects(type="timeseries",
                        key="mykey",
                        tags="tag1 tag2",
                        owner="myuser",
                        app="")

        .. tab:: Async

            ::

                obj = await p.objects(type="timeseries",
                        key="mykey",
                        tags="tag1 tag2",
                        owner="myuser",
                        app="")
        Args:
            type (str): The type of the objects (like "timeseries")
            key (str): The unique app key of the object (each app can only have one object with a given key)
            tags (str): The tags that the object must have, separated by spaces
            owner (str): The owner username of the object (set automatically when accessed using the objects property of a user)
            app (str): The app ID that the object belongs to. Set to empty string for objects that don't belong to any app. (set automatically when accessing the objects property of an app)
            icon (bool,False): Whether to include the icons of the returned objects' data.
        Returns:
            A list of :class:`~heedy.Object` matching the given constraints.
        Throws:
            HeedyException: If the request fails.
        """
        return super()._call(
            f=lambda x: [registry.getObject(xx, self.session) for xx in x], **kwargs
        )

    def create(self, name: str, meta: Dict = {}, type: str = "timeseries", **kwargs):
        """
        Creates a new object of the given type (timeseries by default).
        Only the first argument, the object name is required.

        .. tab:: Sync

            ::

                obj = app.objects.create("My Timeseries",
                    description="This is my timeseries",
                    icon="fas fa-chart-line",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata"
                    key="myts")

        .. tab:: Async

            ::

                obj = await app.objects.create("My Timeseries",
                    description="This is my timeseries",
                    icon="fas fa-chart-line",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata"
                    key="myts")

        When creating an object for an app, it is useful to give it a `key`.
        Keys are unique per-app, meaning that the app can have only one object with the given key.

        Args:
            name (str,""): The name of the object
            description (str,""): A description of the object
            icon (str,""): The icon of the object, either a base64 urlencoded image or fontawesome/material icon id.
            owner (str): The owner of the object (set automatically when accessed using the objects property of a user or app).
            app (str): The app ID that the object belongs to (set automatically when accessing the objects property of an app).
            meta (dict,{}): The metadata of the object.
            type (str,"timeseries"): The type of the object
            key (str): The unique per-app key of the object
            tags (str,""): The tags assigned to the object, separated by spaces
            owner_scope (str,"*"): The space-separated scopes to give the owner of the object if it is managed by an app.
        Returns:
            The newly created :class:`~heedy.Object`, with its data cached.

        Throws:
            HeedyException: if the object could not be created.

        """
        return super()._create(
            f=lambda x: registry.getObject(x, self.session),
            **{"name": name, "type": type, "meta": meta, **kwargs},
        )
