import json
from urllib.parse import urljoin

# Used for the synchronous session
import requests
import socket
from urllib3.connection import HTTPConnection
from urllib3.connectionpool import HTTPConnectionPool
from requests.adapters import HTTPAdapter


# Used for the asynchronous session
import aiohttp

import urllib.parse

import pprint
from typing import Dict

DEFAULT_URL = "http://localhost:1324"


def q(value):
    """
    Quotes values so they are safe to use as elements of a URL
    """
    return urllib.parse.quote(value, safe="")


class HeedyException(Exception):
    """
    HeedyException is raised when the server returns an error value
    in response to a request.

    You can get the error contents by accessing the :code:"error" property
    and the :code:"error_description" property::

        # This is equivalent to print(myerror)
        print(f"{myerror.error}: {myerror.error_description}")
    """

    def __init__(self, msg):
        self.error = msg["error"]
        self.error_description = msg["error_description"]

    def __str__(self):
        return f"{self.error}: {self.error_description}"


class Session:
    """
    Session is the abstract base class that both sync and async sessions implement
    """

    def __init__(self, namespace, url=DEFAULT_URL):
        self.namespace = namespace

        # Set up the API url
        if not url.startswith("unix:"):
            if url.startswith(":"):
                # No host was given, let's use localhost
                url = "localhost" + url
            if not url.startswith("http"):
                url = "https://" + url
            if not url.endswith("/"):
                url = url + "/"
        self.url = url

    def f(self, x,func):
        raise NotImplementedError()

    @property
    def isasync(self):
        raise NotImplementedError()

    def setAccessToken(self, token):
        raise NotImplementedError()

    def setPluginKey(self, key):
        raise NotImplementedError()

    def setHeader(self, key, value):
        raise NotImplementedError()

    def version(self):
        raise NotImplementedError()

    def get(self, path, params={}, f=lambda x: x):
        raise NotImplementedError()

    def post(self, path, data, params={}, f=lambda x: x):
        raise NotImplementedError()

    def patch(self, path, data, params={}, f=lambda x: x):
        raise NotImplementedError()

    def delete(self, path, params={}, f=lambda x: x):
        raise NotImplementedError()

    def close(self):
        raise NotImplementedError()


class UnixConnection(HTTPConnection):
    def __init__(self, sockloc):
        super().__init__("localhost")
        self.sockloc = sockloc

    def connect(self):
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.connect(self.sockloc)


class UnixConnectionPool(HTTPConnectionPool):
    def __init__(self, sockloc):
        super().__init__("localhost")
        self.sockloc = sockloc

    def _new_conn(self):
        return UnixConnection(self.sockloc)


class UnixAdapter(HTTPAdapter):
    def __init__(self, sockloc):
        super().__init__()
        self.sockloc = sockloc

    def get_connection(self, url, proxies=None):
        return UnixConnectionPool(self.sockloc)


class SyncSession(Session):
    """
    SyncSession is to be used in synchronous programs. It uses requests internally.
    """

    def __init__(self, namespace, url=DEFAULT_URL):
        super().__init__(namespace, url)
        self.s = requests.Session()
        self.s.headers.update({"Content-Type": "application/json"})
        if url.startswith("unix:"):
            self.url = "http://unixsocket/"
            self.s.mount(self.url, UnixAdapter(url[5:]))

    def f(self, x, func):
        return func(x)

    @property
    def isasync(self):
        return True

    def setAccessToken(self, token):
        self.s.headers.update({"Authorization": f"Bearer {token}"})

    def setPluginKey(self, key):
        self.s.headers.update({"X-Heedy-Key": key})

    def setHeader(self, key, value):
        self.s.headers.update({key: value})

    def version(self):
        return self.handleResponse(
            self.s.get(urljoin(self.url, "api/server/version"))
        ).text

    def handleResponse(self, r):
        if r.status_code >= 400:
            # The response returned an error
            try:
                msg = r.json()
            except:
                msg = {
                    "error": "malformed_response",
                    "error_description": f'The server returned "{r.text}", which is not json.',
                }
            raise HeedyException(msg)
        return r

    def get(self, path, params={}, f=lambda x: x):
        return f(
            self.handleResponse(
                self.s.get(urljoin(self.url, path), params=params)
            ).json()
        )

    def post(self, path, data, params={}, f=lambda x: x):
        return f(
            self.handleResponse(
                self.s.post(
                    urljoin(self.url, path), data=json.dumps(data), params=params
                )
            ).json()
        )

    def patch(self, path, data, params={}, f=lambda x: x):
        return f(
            self.handleResponse(
                self.s.patch(
                    urljoin(self.url, path), data=json.dumps(data), params=params
                )
            ).json()
        )

    def delete(self, path, params={}, f=lambda x: x):
        return f(
            self.handleResponse(
                self.s.delete(urljoin(self.url, path), params=params)
            ).json()
        )

    def close(self):
        self.s.close()


class AsyncSession(Session):
    """
    AsyncSession is used when running in an asyncio event loop. All of the requests become coroutines,
    allowing them to be awaited
    """

    def __init__(self, namespace, url=DEFAULT_URL):
        super().__init__(namespace, url)
        self.s = None
        self.headers = {"Content-Type": "application/json"}

    @property
    def isasync(self):
        return True

    async def f(self, x, func):
        return func(await x)

    @staticmethod
    def __p(p):
        for k in p:
            if not isinstance(p[k], str):
                p[k] = json.dumps(p[k])
        return p

    def setAccessToken(self, token):
        self.headers["Authorization"] = f"Bearer {token}"

    def setPluginKey(self, key):
        self.headers["X-Heedy-Key"] = key

    def setHeader(self, key, value):
        self.headers[key] = value

    def initSession(self):
        if self.s is None:
            if self.url.startswith("unix:"):
                conn = aiohttp.UnixConnector(path=self.url[5:])
                self.s = aiohttp.ClientSession(connector=conn)
                self.url = "http://unixsocket/"
            else:
                self.s = aiohttp.ClientSession()

    async def handleResponse(self, r):
        if r.status >= 400:
            # The response returned an error
            try:
                msg = await r.json()
            except:
                msg = {
                    "error": "malformed_response",
                    "error_description": f"The server did not return valid json",
                }
            raise HeedyException(msg)
        return r

    async def version(self):
        self.initSession()
        return await (
            await self.handleResponse(
                await self.s.get(urljoin(self.url, "api/server/version"))
            )
        ).text()

    async def get(self, path, params={}, f=lambda x: x):
        self.initSession()
        return f(
            await (
                await self.handleResponse(
                    await self.s.get(
                        urljoin(self.url, path),
                        params=self.__p(params),
                        headers=self.headers,
                    )
                )
            ).json()
        )

    async def post(self, path, data, params={}, f=lambda x: x):
        self.initSession()
        return f(
            await (
                await self.handleResponse(
                    await self.s.post(
                        urljoin(self.url, path),
                        params=self.__p(params),
                        data=json.dumps(data),
                        headers=self.headers,
                    )
                )
            ).json()
        )

    async def patch(self, path, data, params={}, f=lambda x: x):
        self.initSession()
        return f(
            await (
                await self.handleResponse(
                    await self.s.patch(
                        urljoin(self.url, path),
                        params=self.__p(params),
                        data=json.dumps(data),
                        headers=self.headers,
                    )
                )
            ).json()
        )

    async def delete(self, path, params={}, f=lambda x: x):
        self.initSession()
        return f(
            await (
                await self.handleResponse(
                    await self.s.delete(
                        urljoin(self.url, path),
                        params=self.__p(params),
                        headers=self.headers,
                    )
                )
            ).json()
        )

    async def raw(self, method, path, data=None, params={}, headers={}):
        self.initSession()
        return await self.s.request(
            method,
            urljoin(self.url, path),
            headers={**self.headers, **headers},
            data=data,
            params=params
        )

    async def close(self):
        if self.s is not None:
            await self.s.close()


def getSessionType(sessionType: str, namespace: str, url: str = DEFAULT_URL) -> Session:
    """
    This function is given a string, either "sync" or "async", and it returns a SyncSession or AsyncSession respectively.
    """
    if sessionType == "sync":
        return SyncSession(namespace, url)
    if sessionType == "async":
        return AsyncSession(namespace, url)
    raise NotImplementedError(f"The session type '{sessionType}' is not implemented")


from .notifications import Notifications


class APIObject:
    """
    APIObject represents an object in heedy (user,app,object,etc).
    It is given a session and the api location of the object, and allows
    reading, updating, and deleting the object
    """

    props = {"name", "description", "icon"}
    """
    Each element has the above properties available as attributes.
    In synchronous sessions, they allow you to update the properties directly::

        o.name = "My new name"
        assert o.name == "My new name"

    The above is equivalent to::

        o.update(name="My new name")
        assert o["name"] == "My new name"

    Note that each time you access the properties, they are fetched from the server.
    In non-interactive scripts it is useful to avoid redundant querying,
    so each read of data is cached. To use this cached data, you can access
    the property as a key. Instead of :code:`o.name`, use :code:`o["name"]`,
    and call :code:`o.read()` to create/update the cache. 
    Accessing :code:`o["name"]` will only work after the data was initially read,
    otherwise it will lead to a :code:`KeyError`.

    """

    read_qparams = {"icon": {"icon": True}}

    def __init__(self, uri: str, constraints: Dict, session: Session, cached_data=None):
        self.session = session
        self.uri = uri
        self.cached_data = cached_data if cached_data is not None else {}

        #: A :class:`~heedy.notifications.Notifications` object that allows you to access the notifications
        #: associated with this element. See :ref:`python_notifications` for details.
        self.notifications = Notifications(constraints, self.session)

    def read(self, **kwargs):
        """
        Sends a GET request to the element's URI with function arguments as query parameters.
        This method is available for all subclasses of APIObject (objects, apps, users, etc),
        and is used to read element's properties in heedy.

        .. tab:: Sync

            ::

                data = o.read(icon=True)

        .. tab:: Async

            ::

                data = await o.read(icon=True)

        Caches the result of the read accessible as dict keys::

            assert data["name"] == o["name"]

        The read or update functions both update the cached data automatically.

        Args:
            **kwargs: The url parameters to send with the request.
        Returns:
            The server's response dict, namely a dict of the element's properties.
        Raises:
            HeedyException: If the server returns an error.
        """

        def writeCache(o):
            if not "error" in o:
                # This is an update, because we want to keep caching old data
                # that was not requested by this read (for example icon)
                self.cached_data.update(o)
            return o

        return self.session.f(self.session.get(self.uri, params=kwargs), writeCache)

    def update(self, **kwargs):
        """
        Sends a PATCH request to element's URI with arguments as a json object.
        This method is available for all subclasses of APIObject (objects, apps, users, etc),
        and is used to update the element's properties in heedy.

        .. tab:: Sync

            ::

                o.update(name="My new name",description="my new description")
                assert o["name"] == "My new name"

        .. tab:: Async

            ::

                await o.update(name="My new name",description="my new description")
                assert o["name"] == "My new name"

        Args:
            **kwargs: The properties to update, sent as the json body of the request.
        Returns:
            The server's response as a dict, namely the updated element's properties.
        Raises:
            HeedyException: If the server returns an error, such as when there are insufficient permissions.
        """

        def updateCache(o):
            if "result" in o and o["result"] == "ok":
                self.cached_data.update(kwargs)
            return o

        return self.session.f(self.session.patch(self.uri, kwargs), updateCache)

    def delete(self, **kwargs):
        """
        Calls the element's URI with the DELETE method.
        This is a method available for all subclasses of APIObject
        (objects, apps, users, etc), and removes all associated data from Heedy.

        .. tab:: Sync

            ::

                o.delete()
                o.read() # Throws error - it longer exists!

        .. tab:: Async

            ::

                await o.delete()
                await o.read() # Throws error - it no longer exists!
        Args:
            **kwargs: Arguments to pass as query parameters to the server (usually empty)
        Raises:
            HeedyException: If the server returns an error, or when the app does not have permission to delete.
        """
        return self.session.delete(self.uri, params=kwargs)

    def __setattr__(self, name, value):
        if name in self.props:
            return self.update(**{name: value})
        return super().__setattr__(name, value)

    def __getattr__(self, attr: str):
        if attr.startswith("_"):  # ipython tries a bunch of repr formats
            raise AttributeError(f"Unknown attribute '{attr}'")
        qparams = self.read_qparams.get(attr, {})
        return self.session.f(self.read(**qparams), lambda x: x[attr])

    def __eq__(self, other):
        if isinstance(other, self.__class__):
            return other.uri == self.uri
        return False

    def __getitem__(self, i):
        # Gets the item from the cache - assumes that the data is in the cache. If not, need to call .read() first
        return self.cached_data[i]

    def __setitem__(self,key,value):
        if key in self.props:
            return self.update(**{key: value})
        raise KeyError(f"{key} is not a valid property")

    def __str__(self):
        return self.__class__.__name__ + pprint.pformat(self.cached_data)

    def __repr__(self):
        return str(self)

    def notify(self, *args, **kwargs):
        """
        Shorthand for :code:`self.notifications.notify` (see :ref:`python_notifications`).
        """
        return self.notifications.notify(*args, **kwargs)


class APIList:
    """
    APIList is an internal backend class which is used to represent
    a list of objects in heedy (users,apps,objects,etc).
    """

    def __init__(self, uri: str, constraints: Dict, session: Session):
        self.session = session
        self.uri = uri
        self._constraints = constraints

    # These are internal functions that help with implementing the useful parts of
    # lists
    def _create(self, f=lambda x: x, **kwargs):
        return self.session.post(self.uri, {**self._constraints, **kwargs}, f=f)

    def _getitem(self, item, f=lambda x: x):
        return self.session.get(f"{self.uri}/{item}", f=f)

    def _call(self, f=lambda x: x, **kwargs):
        return self.session.get(self.uri, params={**self._constraints, **kwargs}, f=f)
