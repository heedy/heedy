from aiohttp import web

from heedy import Plugin

p = Plugin()

routes = web.RouteTableDef()


@routes.post("/cupdate")
async def index(request):
    print("REQUEST", request.headers)
    print("GOT", await request.json())
    return web.Response(text="OK")


@routes.post("/supdate")
async def index(request):
    print("REQUEST", request.headers)
    print("GOT", await request.json())
    return web.Response(text="OK")


app = web.Application()
app.add_routes(routes)


# Runs the server over a unix domain socket. The socket is automatically placed in the data folder,
# and not the plugin folder.
web.run_app(app, path=f"{p.name}.sock")
