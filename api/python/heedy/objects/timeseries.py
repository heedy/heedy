from .objects import Object
from .registry import registerObjectType

import time
import datetime
import json
from urllib.parse import urljoin

from typing import Union

# Allows querying by string times
from dateparser import parse


def parseTime(t):
    if isinstance(t, str):
        tnew = parse(t)
        if tnew is not None:
            t = tnew
    if isinstance(t, datetime.datetime):
        t = t.timestamp()
    return t


def fixTimestamps(query):
    if "t1" in query:
        query["t1"] = parseTime(query["t1"])
    if "t2" in query:
        query["t2"] = parseTime(query["t2"])
    if "t" in query:
        query["t"] = parseTime(query["t"])


class DatapointArray(list):
    """
    The DatapointArray is a convenience wrapper on data returned from timeseries.
    It allows a bit of extra functionality to make working with timeseries simpler.
    """

    def __init__(self, data: Union[list, str] = []):
        """ Initializes the datapoint array. If given a filename, loads the data from the file"""
        if isinstance(data, str):
            list.__init__(self, [])
            self.load(data)
        else:
            list.__init__(self, data)

    def __add__(self, other):
        return DatapointArray(self).merge(other)

    def __radd__(self, other):
        return DatapointArray(self).merge(other)

    def __getitem__(self, key):
        if key == "t":
            return self.t()
        if key == "d":
            return self.d()
        d = list.__getitem__(self, key)
        if isinstance(key, slice):
            d = DatapointArray(d)
        return d

    def sort(self, f=lambda d: d["t"]):
        """Sort here works by sorting by timestamp by default"""
        list.sort(self, key=f)
        return self

    def d(self):
        """Returns just the data portion of the datapoints as a list"""
        return list(map(lambda x: x["d"], self.raw()))

    def t(self):
        """Returns just the timestamp portion of the datapoints as a list.
        The timestamps are in python datetime's date format."""
        return list(map(lambda x: datetime.datetime.fromtimestamp(x["t"]), self.raw()))

    def dt(self):
        """Returns just the durations of all datapoints."""
        return list(map(lambda x: 0.0 if not "dt" in x else x["dt"], self.raw()))

    def merge(self, array):
        """Adds the given array of datapoints to the generator.
        It assumes that the datapoints are formatted correctly for heedy, meaning
        that they are in the format::

            [{"t": unix timestamp, "d": data}]

        The data does NOT need to be sorted by timestamp - this function sorts it for you
        """
        self.extend(array)
        self.sort()

        return self

    def raw(self):
        """Returns array as a raw python array. For cases where for some reason
        the DatapointArray wrapper does not work for you

        """
        return list.__getitem__(self, slice(None, None))

    def write(self, filename):
        """Writes the data to the given file::

            DatapointArray([{"t": unix timestamp, "d": data}]).writeJSON("myfile.json")

        The data can later be loaded using loadJSON.
        """
        with open(filename, "w") as f:
            json.dump(self, f)

    def load(self, filename):
        """Adds the data from a JSON file. The file is expected to be in datapoint format::

            d = DatapointArray().loadJSON("myfile.json")

        Can be used to read data dumped by writeJSON.
        """
        with open(filename, "r") as f:
            self.merge(json.load(f))
        return self

    def tshift(self, t):
        """Shifts all timestamps in the datapoint array by the given number of seconds.
        It is the same as the 'tshift' pipescript transform.

        Warning: The shift is performed in-place! This means that it modifies the underlying array::

            d = DatapointArray([{"t":56,"d":1}])
            d.tshift(20)
            print(d) # [{"t":76,"d":1}]
        """
        raw = self.raw()
        for i in range(len(raw)):
            raw[i]["t"] += t
        return self

    def sum(self):
        """Gets the sum of the data portions of all datapoints within"""
        raw = self.raw()
        s = 0
        for i in range(len(raw)):
            s += raw[i]["d"]
        return s

    def mean(self):
        """Gets the mean of the data portions of all datapoints within"""
        return self.sum() / float(len(self))

    def pd(self):
        """Returns the data as a pandas dataframe"""
        import pandas

        df = pandas.json_normalize(self)
        df["t"] = pandas.to_datetime(df["t"], unit="s")
        if "dt" in df:
            df["dt"] = pandas.to_timedelta(df["dt"], unit="s")
        return df


class Timeseries(Object):
    def __call__(self, **kwargs):
        """
        Gets timeseries data. You can query by index with i1 and i2, or by timestamp by t1 and t2.
        Timestamps can be strings such as "last month", "1pm" or "jun 5, 2019, 1pm", which will be
        parsed and converted to the corresponding unix timestamps
        """
        fixTimestamps(kwargs)
        return self.session.get(
            self.uri + "/timeseries", params=kwargs, f=lambda x: DatapointArray(x)
        )

    def __getitem__(self, getrange):
        """Allows accessing the timeseries just as if it were just one big python array.
        An example::

            #Returns the most recent 5 datapoints from the timeseries
            timeseries[-5:]

            #Returns all the data the timeseries holds.
            timeseries[:]

        In order to perform transforms on the timeseries and to aggregate data, look at __call__,
        which allows getting index ranges along with a transform.
        """
        if not isinstance(getrange, slice):
            # Return the single datapoint
            return self(i1=getrange, i2=getrange + 1)[0]

        # The query is a slice - return the range
        qkwargs = {}
        if getrange.start is not None:
            qkwargs["i1"] = getrange.start
        if getrange.stop is not None:
            qkwargs["i2"] = getrange.stop
        return self(**qkwargs)

    def length(self):
        """Returns the number of datapoints in the timeseries"""
        return self.session.get(self.uri + "/timeseries/length")

    def insert_array(self, datapoint_array, **kwargs):
        """given an array of datapoints, inserts them to the timeseries. This is different from append(),
        because it requires an array of valid datapoints, whereas append only requires the data portion
        of the datapoint, and fills out the rest::

            s.insert_array([{"d": 4, "t": time.time()},{"d": 5, "t": time.time(), "dt": 5.3}])

        Each datapoint can optionally also contain a "dt" parameter with the datapoint's duration in seconds.
        A time series can't have multiple datapoints with the same timestamp, so such datapoints are automatically
        overwritten by default. Using method="insert" will throw an error if a timestamp conflicts with an existing one.
        """
        return self.session.post(
            self.uri + "/timeseries", data=datapoint_array, params=kwargs
        )

    def append(self, data, duration=0):
        """inserts one datapoint with the given data, and appends it to
        the timeseries, using the current timestamp::

            s.append("Hello World!")

        """
        return self.insert_array([{"d": data, "t": time.time(), "dt": duration}])

    def insert(self, data, timestamp=None, duration=0):
        if timestamp is None:
            return self.append(data, duration)
        return self.insert_array([{"t": timestamp, "d": data, "dt": duration}])

    def remove(self, **kwargs):
        """
        Removes the given data from the timeseries
        """
        return self.session.delete(self.uri + "/timeseries", params=kwargs)

    def save(self, filename):
        """Saves the entire timeseries data to the given filename::

        ts.save("myts.json")


        """
        return self.session.get(
            self.uri + "/timeseries", f=lambda x: DatapointArray(x).writeJSON(filename)
        )

    def load(self, filename, **kwargs):
        """Loads array data from the given file to the timeseries::

        ts.load("myts.json")

        """
        return self.insert_array(DatapointArray().loadJSON(filename), **kwargs)

    def __len__(self):
        return self.length()


registerObjectType("timeseries", Timeseries)
