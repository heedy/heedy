from heedy import Plugin
from aiohttp import web

p = Plugin()

routes = web.RouteTableDef()


@routes.get("/api/hello")
async def hello(request):
    if p.isUser(request):
        user = await p.users[request.headers["X-Heedy-As"]]
        print("Saying hello to", user)
        name = user["name"]
        if name == "":
            name = user["username"]

        return web.Response(text=f"Hello {name}!")
    else:
        return web.Response(text="Hello World!")


app = web.Application()
app.add_routes(routes)

# Runs the server over a unix domain socket. The socket is automatically placed in the heedy data folder,
# and not the plugin folder.
web.run_app(app, path=f"{p.name}.sock")
