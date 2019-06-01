import json
import aiohttp
from aiohttp import web
from urllib.parse import urljoin
import multidict
import os

routes = web.RouteTableDef()

# All info necessary to access heedy is piped in as json to the executable
plugin_configuration = json.loads(input())

# The API key to use when the plugin wants admin access to the heedy api.
apikey = plugin_configuration["apikey"]
overlay = plugin_configuration["overlay"]

# Change local directory
os.chdir(plugin_configuration["data_dir"])


@routes.get("/api/restapi/hello")
async def index(request):
    print("REQUEST", request)
    return web.Response(text="Hello World!")


@routes.get("/api/restapi/hello2")
async def fwd(request):
    print("GOT REQUEST, Forwarding")

    target_url = urljoin("http://localhost:1324", request.path)
    print(target_url, request.headers)

    data = await request.read()
    get_data = request.rel_url.query
    print(data, get_data)

    headers = multidict.CIMultiDict(request.headers)

    headers["X-Heedy-Key"] = apikey
    headers["X-Heedy-Auth"] = "public"
    headers["X-Heedy-Overlay"] = str(overlay)

    print(headers)

    async with aiohttp.ClientSession(
        connector=aiohttp.TCPConnector(ssl=False)
    ) as session:
        async with session.request(
            request.method, target_url, headers=headers, data=data
        ) as resp:
            res = resp
            raw = await res.read()

    return web.Response(text="p2bitches:" + raw.decode())


app = web.Application()
app.add_routes(routes)

# Runs the server over a unix domain socket. The socket is automatically placed in the data folder,
# and not the plugin folder.
web.run_app(app, path="p2.sock")

