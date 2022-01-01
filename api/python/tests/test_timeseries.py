import pytest
from heedy import App, Timeseries, DatapointArray
import pandas as pd


def test_ts():
    a = App("testkey")
    ts = a.objects.create("myts", key="key1")
    assert len(ts) == 0
    ts.insert("hi!")
    assert len(ts) == 1
    assert len(ts(t1="now-10s")) == 1
    assert len(ts(t1="now")) == 0
    assert ts[0]["d"] == "hi!"
    assert ts[-1]["d"] == "hi!"
    ts.insert("hello")
    assert ts[0]["d"] == "hi!"
    assert ts[-1]["d"] == "hello"
    ts.delete()


def test_df():
    a = App("testkey")
    ts = a.objects.create("myts", key="key1")
    assert len(ts) == 0
    assert isinstance(ts[:].to_df(), pd.DataFrame)

    ts.insert(5)
    ts.insert(6)

    assert isinstance(ts[:].to_df(), pd.DataFrame)
    assert len(ts[:].to_df().d) == 2
    assert ts[:].to_df().d[0] == 5
    assert ts[:].to_df().d[1] == 6

    assert isinstance(ts[:], DatapointArray)
    assert isinstance(ts(output_type="dataframe"), pd.DataFrame)
    Timeseries.output_type = "dataframe"
    assert isinstance(ts[:], pd.DataFrame)
    assert isinstance(ts(output_type="list"), DatapointArray)

    ts.delete()
