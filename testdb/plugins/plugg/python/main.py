import json
from aiohttp import web

apikey = input()
config = json.loads(input())

print("APIKEY:", apikey)
print("CONFIG:", config)


routes = web.RouteTableDef()


@routes.get("/scree")
async def index(request):
    print("GOT REQUEST plugg")
    return web.Response(text="weyaoooo")


app = web.Application()
app.add_routes(routes)
web.run_app(app, port=config["plugin"]["plugg"]["exec"]["syncer"]["port"])

