import pytest
from heedy import App,Plugin, HeedyException
import requests

plugin_config = requests.get("http://localhost:1324/api/testplugin").json()

def test_pluginbasics():
    p = Plugin(plugin_config,session="sync")

    u = p.users()

    assert len(u)==1
    assert u[0].username=="test"


def test_apps():
    p = Plugin(plugin_config,session="sync")

    a = p.apps()

    assert len(a)==1 # test app, leave it alone for now

    app = p.apps.create("My App Name",
        owner="test",
        plugin=f"{p.name}:myapp"
    )

    app2 = p.users["test"].apps.create("My App Name",
        plugin=f"{p.name}:myapp"
    )

    assert len(p.apps())==3
    assert len(p.apps(owner="test"))==3
    assert len(p.apps(plugin=f"{p.name}:myapp"))==2

    app.delete()
    app2.delete()
    assert len(p.apps())==1

    with pytest.raises(HeedyException):
        app = p.apps.create("My App Name",
            owner="test",
            plugin=f"notplugin:myapp"
        )
    with pytest.raises(HeedyException):
        app = p.apps.create("My App Name",
            owner="test",
            plugin="k"
        )

    # Test creation of app that is pre-defined in heedy config
    app = p.apps.create(plugin="testplugin:testapp",owner="test")
    assert app["name"]=="Test App"
    assert app.description=="Hello World"

    objs = app.objects()
    assert len(objs)==1
    assert objs[0]["key"] == "foobar"

    app.delete()

    app = p.apps.create(plugin="testplugin:testapp",owner="test",description="hello2")
    assert app["name"]=="Test App"
    assert app.description=="hello2"
    app.delete()

    a = p.apps()
    assert len(a)==1 # test app, leave it alone for now


@pytest.mark.asyncio
async def test_apps_async():
    p = Plugin(plugin_config)

    a = await p.apps()

    assert len(a)==1