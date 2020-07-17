import pytest
from heedy import App


def test_ts():
    a = App("testkey")
    ts = a.objects.create("myts", key="key1")
    assert len(ts) == 0
    ts.insert("hi!")
    assert len(ts) == 1
    assert len(ts(t1="now-10s")) == 1
    assert len(ts(t1="now")) == 0
