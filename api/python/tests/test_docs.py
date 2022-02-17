import pytest
from heedy import App
import time

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