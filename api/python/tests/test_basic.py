import pytest
from heedy import App


def test_basics():
    a = App("testkey")
    assert a.owner.username == "test"

    a.owner.name = "Myname"
    assert a.owner.name == "Myname"

    assert len(a.objects()) == 0

    o = a.objects.create("myobj", {"schema": {"type": "number"}})
    assert o.name == "myobj"
    assert o.type == "timeseries"
    assert len(a.objects()) == 1

    assert o == a.objects[o.id]

    assert o.length() == 0
    o.append(2)
    assert o.length() == 1
    d = o[:]
    assert len(d) == 1
    assert d[0]["d"] == 2
    assert "dt" not in d[0]
    o.append(3, duration=9)
    d = o[:]
    assert len(d) == 2
    assert d[1]["d"] == 3
    assert d[1]["dt"] == 9
    o.remove()  # Clear the timeseries
    assert len(o) == 0

    o.delete()
    assert len(a.objects()) == 0
    # assert len(a.owner.apps())==1


def test_notifications():
    a = App("testkey")

    assert len(a.notifications()) == 0
    a.notify("hi", "hello")
    assert len(a.notifications()) == 1
    a.notifications.delete("hi")
    assert len(a.notifications()) == 0


def test_kv():
    a = App("testkey")

    assert len(a.kv()) == 0
    a.kv["test"] = True
    assert len(a.kv()) == 1
    assert a.kv["test"] == True

    del a.kv["test"]
    assert len(a.kv()) == 0

def test_tags_and_keys():
    a = App("testkey")
    a.objects.create("obj1",tags="tag1 tag2",key="key1")
    a.objects.create("obj2",tags="tag1 tag3",key="key2")

    assert len(a.objects(key="key1"))==1
    assert len(a.objects(key="key2"))==1
    assert len(a.objects(key="key3"))==0
    assert len(a.objects(key=""))==0

    assert len(a.objects(tags="tag1"))==2
    assert len(a.objects(tags="tag1 tag3"))==1
    assert len(a.objects(tags="tag1 tag3",key="key1"))==0

    with pytest.raises(Exception):
        a.objects.create("obj3",tags="tag1 tag3",key="key2")

    a.objects(key="key1")[0].tags = "tag4"
    assert len(a.objects(tags="tag1"))==1
    assert len(a.objects(tags="tag4"))==1

    a.objects(key="key2")[0].key = ""
    assert len(a.objects(key="key2"))==0
    assert len(a.objects(key=""))==1

    for o in a.objects():
        o.delete()



@pytest.mark.asyncio
async def test_basics_async():
    a = App("testkey", session="async")

    await (await a.owner).update(name="Myname2")
    assert await (await a.owner).name == "Myname2"

    assert len(await a.objects()) == 0

    o = await a.objects.create("myobj2", {"schema": {"type": "number"}})
    assert o.name == "myobj2"
    assert o.type == "timeseries"
    assert len(await a.objects()) == 1

    assert o == await a.objects[o.id]

    await o.delete()

    # assert len(await (await a.owner).apps())==1
