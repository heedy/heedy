import json
import os
import aiohttp
from aiohttp import web
from urllib.parse import urljoin
import multidict


routes = web.RouteTableDef()

# All info necessary to access heedy is piped in as json to the executable
plugin_configuration = json.loads(input())
print("unix_revproxy")

# Change local directory
os.chdir(plugin_configuration["data_dir"])

# The API key to use when the plugin wants admin access to the heedy api.
apikey = plugin_configuration["apikey"]


@routes.get("/test/me")
async def index(request):
    return web.Response(text="Hello World!")


app = web.Application()
app.add_routes(routes)

os.makedirs("test", exist_ok=True)

# Runs the server over a unix domain socket. The socket is automatically placed in the data folder,
# and not the plugin folder.
web.run_app(app, path="test/server.sock")

