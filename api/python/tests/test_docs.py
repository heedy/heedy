import pytest
from heedy import App, Plugin, HeedyException
import time
import requests
import aiohttp

def test_objects():
    app = App("testkey")
    assert len(app.objects())==0

    obj = app.objects.create("My Timeseries",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata",
                    key="myts")

    objs = app.objects()
    assert len(objs) == 1
    assert objs[0] == obj
    assert objs[0].id==obj.id
    assert objs[0].meta == obj.meta
    assert objs[0]["name"] == "My Timeseries"
    assert objs[0]["type"] == "timeseries"
    assert obj.name == "My Timeseries"
    
    # Reading and Updating properties

    # Reads the key property from the server
    assert obj.key == "myts"
    # Uses the previously read cached data, avoiding a server query
    assert obj["key"] == "myts"

    obj.read()
    assert obj["key"]=="myts"

    obj.description = "My description"
    assert obj.description == "My description"
    obj.update(description="My description2")
    assert obj["description"] == "My description2"
    obj["description"] = "My description3"
    assert obj.description == "My description3"

    assert obj.meta == {"schema": {"type": "number"}}
    assert obj.meta() == {"schema": {"type": "number"}}

    obj.meta.schema = {"type": "string"}
    assert obj.meta == {"schema": {"type": "string"}}
    assert obj.meta() == {"schema": {"type": "string"}}
    obj.meta["schema"] = {"type": "boolean"}
    assert obj.meta == {"schema": {"type": "boolean"}}
    del obj.meta.schema
    assert obj.meta() == {"schema": {}}


    assert app.objects(key="myts")[0] == obj

    with pytest.raises(Exception):
        assert obj["app"]==app
    with pytest.raises(Exception):
        assert obj.app == app

    app.read()
    assert obj["app"]==app
    assert obj.app == app

    obj.delete()
    assert len(app.objects())==0


@pytest.mark.asyncio
async def test_objects_async():
    app = App("testkey",session="async")
    objs = await app.objects()
    assert len(objs) == 0

    obj = await app.objects.create("My Timeseries",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata",
                    key="myts")
    objs = await app.objects()
    assert len(objs) == 1
    assert objs[0] == obj
    assert objs[0].id==obj.id
    assert objs[0].meta == obj.meta
    assert objs[0]["name"] == "My Timeseries"
    assert objs[0]["type"] == "timeseries"
    assert (await (obj.name)) == "My Timeseries"


    # Reading and Updating Properties
    # Reads the key property from the server
    assert (await obj.key) == "myts"
    # Uses the previously read cached data, avoiding a server query
    assert obj["key"] == "myts"

    await obj.read()
    assert obj["key"]=="myts"

    await obj.update(description="My description")
    assert obj["description"] == "My description"

    assert obj.meta == {"schema": {"type": "number"}}
    assert (await obj.meta()) == {"schema": {"type": "number"}}

    await obj.meta.update(schema={"type": "string"})
    assert obj.meta == {"schema": {"type": "string"}}
    assert (await obj.meta()) == {"schema": {"type": "string"}}
    await obj.meta.delete("schema")
    assert (await obj.meta()) == {"schema": {}}

    assert (await app.objects(key="myts"))[0] == obj


    with pytest.raises(Exception):
        assert obj["app"]==app
    with pytest.raises(Exception):
        assert (await obj.app) == app
    await app.read()
    assert obj["app"]==app
    assert (await obj.app) == app

    await obj.delete()
    assert len(await (app.objects()))==0

def test_ts():
    app = App("testkey")

    obj = app.objects.create("My Timeseries",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata",
                    key="myts")

    assert len(obj)==0
    
    obj.append(5)
    assert obj[-1]["d"]==5

    # Insert the given array of data
    obj.insert_array([{"d": 6, "t": time.time()},{"d": 7, "t": time.time()+0.01, "dt": 5.3}])

    assert len(obj)==3
    assert obj[-1]["d"]==7
    assert obj[1:].d()==[6,7]
    assert obj[1:].d()==[6,7]

    ts = obj(t1="now-1d")
    assert len(ts)==3

    ts = obj(t1="now-1d",t2="now-10s")
    assert len(ts)==0

    ts = obj(t1="1 hour ago")
    assert len(ts)==3

    assert obj(transform="sum")[0]["d"]==18

    obj.remove(i=-1)
    assert len(obj)==2
    assert obj[-1]["d"]==6

    obj.remove(t=ts[0]["t"])

    assert len(obj)==1
    assert obj[0]["d"]==6


    obj.delete()

@pytest.mark.asyncio
async def test_ts_async():
    app = App("testkey",session="async")

    obj = await app.objects.create("My Timeseries",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata",
                    key="myts")

    assert (await obj.length())==0

    # Add a datapoint 5 with current timestamp to the series
    await obj.append(5)
    assert (await obj(i=-1))[0]["d"]==5

    # Insert the given array of data
    await obj.insert_array([{"d": 6, "t": time.time()},{"d": 7, "t": time.time()+0.01, "dt": 5.3}])

    assert (await obj.length())==3
    assert (await obj(i=-1))[0]["d"]==7
    assert (await obj(i1=1)).d()==[6,7]

    ts = await obj(t1="now-1d")
    assert len(ts)==3

    ts = await obj(t1="now-1d",t2="now-10s")
    assert len(ts)==0

    await obj.remove(i=-1)
    assert (await obj.length())==2
    assert (await obj[-1])["d"]==6

    await obj.delete()

@pytest.mark.order("last")
def test_app_keychange():
    # This test has to happen last, because once the app key is changed,
    # there is no more logging in from other tests.
    app = App("testkey",session="sync")
    result = app.update(name="My Test App")
    assert not "access_token" in result

    result = app.update(access_token=True)

    with pytest.raises(HeedyException):
        app.read()

    assert result["access_token"]!="testkey"
    newapp = App(result["access_token"],url=app.session.url)

    assert newapp.name == "My Test App"

        

def test_app():
    app = App("testkey",session="sync")
    assert app.id == "self"
    app.read()
    assert app.id!="self"
    app.description = "hello worldd"

    assert app["description"] == "hello worldd"


    assert app["owner"]["username"]=="test"
    assert app.owner["username"]=="test"

@pytest.mark.asyncio
async def test_app_async():
    app = App("testkey",session="async")
    assert app.id == "self"
    await app.read()
    assert app.id!="self"
    await app.update(description="hello worldd")

    assert app["description"] == "hello worldd"

    assert app["owner"]["username"]=="test"
    assert (await app.owner)["username"]=="test"


def test_apps():
    plugin_config = requests.get("http://localhost:1324/api/testplugin").json()
    p = Plugin(plugin_config,session="sync")

    a = p.apps()

    assert len(a)==1 # test app, leave it alone for now



@pytest.mark.asyncio
async def test_apps_async():
    async with aiohttp.ClientSession() as session:
        async with session.get("http://localhost:1324/api/testplugin") as resp:
            plugin_config = await resp.json()
    p = Plugin(plugin_config)

    a = await p.apps()

    assert len(a)==1