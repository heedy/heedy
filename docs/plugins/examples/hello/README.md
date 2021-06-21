# My First Plugin

This tutorial will teach you the fundamentals of building a basic "Hello World"-style Heedy plugin. This tutorial assumes that you have `heedy` installed, and are on linux or mac (a raspberry pi will do too!). It assumes basic familiarity with the linux command line.

We will start by setting up a heedy database for development and testing of the plugin, and then create a basic Python server that extends Heedy's API.

## Creating a Testing Database

One big no-no of software engineering is testing code on production systems - writing a plugin can be messy work, and depending on how deeply the plugin integrates with Heedy's backend database, you could accidentally delete your data or corrupt the database if you experiment on your main system.

Instead, we will create a test database in which the plugin will be developed:

```bash
heedy create testdb
```

The above command will start a heedy server on port 1324, and you can then set up a user from the browser - username `test` and password `test` usually does the trick! This will then create a database in the `testdb` folder.

Looking inside the folder, we can see the following directory structure:
```
testdb/
    heedy.conf      # the main heedy configuration file
    data/
        heedy.db    # an sqlite3 database containing *everything*
        heedy.sock  # a unix socket that exists only when heedy is running, exposing the plugin REST API.
```

## Creating a Plugin

Now that we have a heedy database, we will make a basic do-nothing plugin called "hello". We start by creating the plugins directory, since it does not exist yet, then we create a folder for our plugin:
```
mkdir testdb/plugins
mkdir testdb/plugins/hello
```

In order for heedy to recognize the plugin, it will need a basic configuration. We write the following to `testdb/plugins/hello/heedy.conf`:

```javascript
plugin "hello" {
    version="1.0.0"
    description = "Introduction to Heedy plugins!"
    icon="fas fa-hand-sparkles" // Use a fontawesome hand icon for the plugin
    license = "Apache-2.0"
}
```

The plugin name must be identical to the folder name. Once you save this file, the plugin will show up as an option in heedy's server configuration.

This is the minimal viable plugin - it does nothing, but it can be enabled in heedy, either through the web GUI, or by setting `active_plugins=["hello"]` in `testdb/heedy.conf` (the global configuration file for your database).

Once enabled, any changes to the plugin configuration or backend servers (we'll get to that in a moment!) will require a heedy restart to take effect, while changes to frontend javascript will simply require a browser refresh.

Before moving on to adding functionality to the plugin, let's document what it does by adding a `README.md` to the plugin folder, where you can describe what the plugin does using Markdown:
```
# Wow

Such an awesome plugin
```

Heedy will then diplay the README in the plugin description dialog:


## It has to say "Hello!"

A plugin that sits there and does nothing isn't particularly useful, so let's make it say hello. 
Heedy plugins are divided into two separate portions: backend servers and frontend javascript. We will start by adding a backend API service that says hello to whomever calls it.

While these servers can be made in any programming language, Heedy has special built-in support for Python, so we will create a basic server using the [aiohttp](https://docs.aiohttp.org/en/stable/) library.

We create the file `testdb/plugins/hello/hello.py`, and add the following to it:
```python
# You can use any Python server library, but the heedy python library has some quality-of-life
# extras built for aiohttp use.
from aiohttp import web
routes = web.RouteTableDef()

# When starting the plugin server, heedy will send a bunch of important data to it on stdin.
# The Plugin object is meant to be global, and reads this data from input on initialization.
# It also changes the working directory to the `data` folder in the heedy database, so that the plugin can
# access the database or write its own files.
from heedy import Plugin
p = Plugin()

# The plugin will answer GET requests at /api/hello, returning a simple text response.
@routes.get("/api/hello")
async def hello(request):
    return web.Response(text="Hello World!")


# Runs the plugin's backend server
app = web.Application()
app.add_routes(routes)
web.run_app(app, path=f"{p.name}.sock")
```