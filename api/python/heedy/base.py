import json
from urllib.parse import urljoin

# Used for the synchronous session
import requests

# Used for the asynchronous session
import aiohttp

DEFAULT_URL = "http://localhost:1324"


class HeedyError(Exception):
    """
    HeedyError is raised when the server returns an error value
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

    def __init__(self, url=DEFAULT_URL):

        # Set up the API url
        if not url.startswith("http"):
            url = "https://" + url
        if not url.endswith("/"):
            url = url + "/"
        self.url = url

    def setAccessToken(self, token):
        raise NotImplementedError()

    def setPluginKey(self, key):
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


class SyncSession(Session):
    """
    SyncSession is to be used in synchronous programs. It uses requests internally.
    """

    def __init__(self, url=DEFAULT_URL):
        super().__init__(url)
        self.s = requests.Session()
        self.s.headers.update({'Content-Type': 'application/json'})

    def setAccessToken(self, token):
        self.s.headers.update({'Authorization': f"Bearer {token}"})

    def setPluginKey(self, key):
        self.s.headers.update({"X-Heedy-Key": key})

    def version(self):
        return self.handleResponse(self.s.get(urljoin(self.url, "api/heedy/v1/server/version"))).text

    def handleResponse(self, r):
        if r.status_code >= 400:
            # The response returned an error
            try:
                msg = r.json()
            except:
                msg = {"error": "malformed_response", "error_description":
                       f"The server returned \"{r.text}\", which is not json."}
            raise HeedyError(msg)
        return r

    def get(self, path, params={}, f=lambda x: x):
        return f(self.handleResponse(self.s.get(urljoin(self.url, path), params=params)).json())

    def post(self, path, data, params={}, f=lambda x: x):
        return f(self.handleResponse(self.s.post(urljoin(self.url, path), data=json.dumps(data), params=params)).json())

    def patch(self, path, data, params={}, f=lambda x: x):
        return f(self.handleResponse(self.s.patch(urljoin(self.url, path), data=json.dumps(data), params=params)).json())

    def delete(self, path, params={}, f=lambda x: x):
        return f(self.handleResponse(self.s.delete(urljoin(self.url, path), params=params)).json())

    def close(self):
        self.s.close()


class AsyncSession(Session):
    """
    AsyncSession is used when running in an asyncio event loop. All of the requests become coroutines,
    allowing them to be awaited
    """

    def __init__(self, url=DEFAULT_URL):
        super().__init__(url)
        self.s = None
        self.headers = {'Content-Type': 'application/json'}

    def setAccessToken(self, token):
        self.headers['Authorization'] = f"Bearer {token}"

    def setPluginKey(self, key):
        self.headers["X-Heedy-Key"] = key

    def initSession(self):
        if self.s is None:
            self.s = aiohttp.ClientSession()

    async def handleResponse(self, r):
        if r.status >= 400:
            # The response returned an error
            try:
                msg = await r.json()
            except:
                msg = {"error": "malformed_response",
                       "error_description": f"The server did not return valid json"}
            raise HeedyError(msg)
        return r

    async def version(self):
        self.initSession()
        return await (await self.handleResponse(await self.s.get(urljoin(self.url, "api/heedy/v1/server/version")))).text()

    async def get(self, path, params={}, f=lambda x: x):
        self.initSession()
        return f(await (await self.handleResponse(await self.s.get(urljoin(self.url, path), params=params, headers=self.headers))).json())

    async def post(self, path, data, params={}, f=lambda x: x):
        self.initSession()
        return f(await (await self.handleResponse(await self.s.post(urljoin(self.url, path), params=params, data=json.dumps(data), headers=self.headers))).json())

    async def patch(self, path, data, params={}, f=lambda x: x):
        self.initSession()
        return f(await (await self.handleResponse(await self.s.patch(urljoin(self.url, path), params=params, data=json.dumps(data), headers=self.headers))).json())

    async def delete(self, path, params={}, f=lambda x: x):
        self.initSession()
        return f(await (await self.handleResponse(await self.s.delete(urljoin(self.url, path), params=params, headers=self.headers))).json())

    async def raw(self, method, path, data=None, params={}, headers={}):
        self.initSession()
        return await self.s.request(method, urljoin(self.url, path), headers={**self.headers, **headers}, data=data)

    async def close(self):
        if self.s is not None:
            await self.s.close()


def getSessionType(sessionType: str, url: str = DEFAULT_URL) -> Session:
    """
    This function is given a string, either "sync" or "async", and it returns a SyncSession or AsyncSession respectively.
    """
    if sessionType == "sync":
        return SyncSession(url)
    if sessionType == "async":
        return AsyncSession(url)
    raise NotImplementedError(
        f"The session type '{sessionType}' is not implemented")


class APIObject:
    """
    APIObject represents an object in heedy (user,app,object,etc).
    It is given a session and the api location of the object, and allows
    reading, updating, and deleting the object
    """

    def __init__(self, session: Session, uri: str):
        self.session = session
        self.uri = uri

    def read(self, **kwargs):
        """
        Read the object
        """
        return self.session.get(self.uri, params=kwargs)

    def update(self, **kwargs):
        """
        Updates the given data::

            o.update(name="My new name",description="my new description")
        """
        return self.session.patch(self.uri, kwargs)

    def delete(self, **kwargs):
        """
        Deletes the object
        """
        return self.session.delete(self.uri, params=kwargs)
