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

        self.notifications = Notifications({}, self.session)
        self.apps = Apps({}, self.session)
        self.users = Users({}, self.session)
        self.objects = Objects({}, self.session)

    @property
    def name(self):
        return self.config["plugin"]

    def copy(self):
        return Plugin(self.config, "async" if self.session.isasync else "sync")

    def query_as(self, accessor):
        p = self.copy()
        p.session.setHeader("X-Heedy-As", accessor)
        return p

    async def forward(
        self, request, data=None, headers={}, run_as: str = None, overlay=None
    ):
        """
        Forwards the given request to the underlying database.
        It only functions in async mode.

        Returns the response.
        """
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
        return self.notifications.notify(*args, **kwargs)
