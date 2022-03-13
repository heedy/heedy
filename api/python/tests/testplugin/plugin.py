from aiohttp import web
import json
import os
routes = web.RouteTableDef()

# Read the plugin configuration that will be exposed at the API
pluginconfig = input()

configjson = json.loads(pluginconfig)

os.chdir(configjson["data_dir"])

@routes.get("/api/testplugin")
async def response(request):
    return web.Response(body=pluginconfig,content_type="application/json")

# Runs the plugin's backend server
app = web.Application()
app.add_routes(routes)
web.run_app(app, path="testplugin.sock")
