from .base import getSessionType
import json
import os
import multidict
import aiohttp
import base64

from .notifications import Notifications

from .users import Users
from .apps import Apps
from .objects import Objects


class Plugin:
    def __init__(self, config=None, session: str = "async"):
        # Load the plugin configuration
        self.config = config
        if self.config is None:
            self.config = json.loads(input())

            # Change the directory to the data dir
            os.chdir(self.config["data_dir"])

        self.session = getSessionType(session, self.name, self.config["config"]["api"])
        self.session.setPluginKey(self.config["apikey"])

        #: A :class:`~heedy.notifications.Notifications` object that allows you to access
        #: to notifications in the Heedy instance. See :ref:`python_notifications` for details.
        self.notifications = Notifications({}, self.session)

        #: An :class:`~heedy.Apps` instance, allowing
        #: to interact with the users in Heedy.
        #: Listing apps that conform to the given restrictions
        #: can be done by calling ``plugin.apps()``.
        self.apps = Apps({}, self.session)

        #: An :class:`~heedy.Users` instance, allowing
        #: to interact with the users in Heedy.
        #: Users can be listed with ``plugin.objects()``.
        self.users = Users({}, self.session)

        #: An :class:`~heedy.Objects` instance, allowing
        #: to interact with the objects in Heedy. Listing objects
        #: that conform to given restrictions can be done by calling ``plugin.objects()``.
        self.objects = Objects({}, self.session)

    @property
    def name(self):
        return self.config["plugin"]

    def copy(self, session=None):
        if session is None:
            session = "async" if self.session.isasync else "sync"
        return Plugin(self.config, session)

    def query_as(self, accessor):
        p = self.copy()
        p.session.setHeader("X-Heedy-As", accessor)
        return p

    async def forward(
        self, request, data=None, headers=None, overlay=None
    ):
        """
        Forwards the given request to the underlying database.
        It only functions in async mode.

        Returns the response.
        """
        if headers is None:
            headers = {}
        if data is None:
            data = await request.read()
        headers = {**request.headers, **headers}

        if overlay is None:
            overlay = "next"

        headers["X-Heedy-Overlay"] = overlay

        return await self.session.raw(
            request.method,
            request.path,
            headers=headers,
            data=data,
            params=request.query,
        )

    async def respond_forwarded(self, request, **kwargs):
        """
        Responds to the request with the result of forward()
        """
        req_res = await self.forward(request, **kwargs)

        response = aiohttp.web.StreamResponse(
            status=req_res.status, headers=req_res.headers
        )
        await response.prepare(request)
        while True:
            chunk = await req_res.content.read(32768)
            if not chunk:
                break
            await response.write(chunk)
        await response.write_eof()
        return response

    def objectRequest(self, request):
        h = request.headers
        modified_date = h["X-Heedy-Modified-Date"]
        if modified_date == "null":
            modified_date = None
        return {
            "request": h["X-Heedy-Request"],
            "id": h["X-Heedy-Id"],
            "modified_date": modified_date,
            "meta": json.loads(base64.b64decode(h["X-Heedy-Meta"])),
            "object": h["X-Heedy-Object"],
            "type": h["X-Heedy-Type"],
            "owner": h["X-Heedy-Owner"],
            "as": h["X-Heedy-As"],
            "access": h["X-Heedy-Access"].split(" "),
        }

    def hasAccess(self, request, scope):
        a = request.headers["X-Heedy-Access"].split(" ")
        return "*" in a or scope in a

    def isUser(self, request):
        h = request.headers
        return (
            not "/" in h["X-Heedy-As"]
            and h["X-Heedy-As"] != "heedy"
            and h["X-Heedy-As"] != "public"
        )

    def isApp(self, request):
        return "/" in request.headers["X-Heedy-As"]

    def isAdmin(self, request):
        request.headers["X-Heedy-As"] == "heedy"

    def fire(self, event):
        """
        Fires the given event
        """
        return self.session.post("/api/events", event)

    def notify(self, *args, **kwargs):
        """
        Shorthand for :code:`self.notifications.notify` (see :ref:`python_notifications`).
        """
        return self.notifications.notify(*args, **kwargs)
