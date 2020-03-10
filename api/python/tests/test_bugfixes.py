import pytest
from heedy import App

# These tests are of bugfixes to heedy, making sure that things are working


def test_appscope():
    # Makes sure that removing owner's access doesn't affect the App's access
    a = App("testkey")
    o = a.objects.create("myobj")
    assert o.access == "*"
    o.key = "lol"
    o.owner_scopes = "read"
    assert o.read()["key"] == "lol"
    o.key = "hiya"
    assert o.read()["key"] == "hiya"


def test_metamod():
    a = App("testkey")
    o = a.objects.create("myobj", otype="timeseries")
    o.meta = {"schema": {"type": "number"}}
    assert o.cached_data["meta"]["schema"]["type"] == "number"
    assert o.cached_data["meta"]["actor"] == False

    assert o.meta["schema"]["type"] == "number"
    assert o.meta["actor"] == False
