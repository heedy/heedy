from .base import getSessionType
import json
import os
import multidict
import aiohttp


class Plugin:
    def __init__(self, session: str = "async"):
        # Load the plugin configuration
        self.config = json.loads(input())

        # Change the directory to the data dir
        os.chdir(self.config["data_dir"])

        self.session = getSessionType(
            session, f"http://localhost:{self.config['config']['port']}")
        self.session.setPluginKey(self.config["apikey"])

    @property
    def name(self):
        return self.config["plugin"]

    async def forward(self, request, data=None, headers={}, run_as: str = None, overlay=None):
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

        return await self.session.raw(request.method, request.path, headers=headers, data=data, params=request.query)

    async def respond_forwarded(self, request, **kwargs):
        """
        Responds to the request with the result of forward()
        """
        req_res = await self.forward(request, **kwargs)

        response = aiohttp.web.StreamResponse(
            status=req_res.status, headers=req_res.headers)
        await response.prepare(request)
        while True:
            chunk = await req_res.content.read(32768)
            if not chunk:
                break
            await response.write(chunk)
        await response.write_eof()
        return response

    def fire(self, event):
        """
        Fires the given event
        """
        return self.session.post("/api/heedy/v1/events", event)

    def listApps(self, **kwargs):
        return self.session.get("api/heedy/v1/apps", params=kwargs)

    def notify(self, n, **kwargs):
        return self.session.post("/api/heedy/v1/notifications", n, params=kwargs)

    def delete_notification(self, n):
        return self.session.delete("/api/heedy/v1/notifications", params=n)
