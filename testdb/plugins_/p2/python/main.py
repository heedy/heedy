import json
import aiohttp
from aiohttp import web
from urllib.parse import urljoin
import multidict

apikey = input()
config = json.loads(input())

print("APIKEY (P2):", apikey)
print("CONFIG (P2):", config)


routes = web.RouteTableDef()


@routes.get("/scree")
async def index(request):
    print("GOT REQUEST, Forwarding")

    target_url = urljoin("https://localhost:3000", request.path)
    print(target_url, request.headers)

    data = await request.read()
    get_data = request.rel_url.query
    print(data, get_data)

    headers = multidict.CIMultiDict(request.headers)

    headers["X-Cdb-Plugin"] = apikey

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
web.run_app(app, port=config["plugin"]["p2"]["exec"]["syncer"]["port"])

