from .objects import Object
from .registry import registerObjectType

import time
import datetime
import json
import os.path
from urllib.parse import urljoin

# Allows querying by string times
from dateparser import parse


def parseTime(t):
    if isinstance(t, str):
        t = parse(t)
        if t is None:
            raise AttributeError("Could not parse timestamp")
    if isinstance(t, datetime.datetime):
        t = datetime.timestamp(t)
    return t


class DatapointArray(list):
    """
    The DatapointArray is a convenience wrapper on data returned from streams.
    It allows a bit of extra functionality to make working with streams simpler.
    """

    def __init__(self, data=[]):
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
            d._d = self._d[key]
            d._t = self._t[key]
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

    def writeJSON(self, filename):
        """Writes the data to the given file::

            DatapointArray([{"t": unix timestamp, "d": data}]).writeJSON("myfile.json")

        The data can later be loaded using loadJSON.
        """
        with open(filename, "w") as f:
            json.dump(self, f)

    def loadJSON(self, filename):
        """Adds the data from a JSON file. The file is expected to be in datapoint format::

            d = DatapointArray().loadJSON("myfile.json")
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


class Stream(Object):
    def __call__(self, actions=False, **kwargs):
        """
        Gets stream data. You can query by index with i1 and i2, or by timestamp by t1 and t2.
        Timestamps can be strings such as "last month", "1pm" or "jun 5, 2019, 1pm", which will be
        parsed and converted to the corresponding unix timestamps
        """
        if "t1" in kwargs:
            kwargs["t1"] = parseTime(kwargs["t1"])
        if "t2" in kwargs:
            kwargs["t2"] = parseTime(kwargs["t2"])
        if "t" in kwargs:
            kwargs["t"] = parseTime(kwargs["t"])
        urimod = "/data"
        if actions:
            urimod = "/actions"
        return self.session.get(
            self.uri + urimod, params=kwargs, f=lambda x: DatapointArray(x)
        )

    def __getitem__(self, getrange):
        """Allows accessing the stream just as if it were just one big python array.
        An example::

            #Returns the most recent 5 datapoints from the stream
            stream[-5:]

            #Returns all the data the stream holds.
            stream[:]

        In order to perform transforms on the stream and to aggregate data, look at __call__,
        which allows getting index ranges along with a transform.
        """
        if not isinstance(getrange, slice):
            # Return the single datapoint
            return self(i1=getrange, i2=getrange + 1)[0]

        # The query is a slice - return the range
        return self(i1=getrange.start, i2=getrange.stop)

    def length(self, actions=False):
        urimod = "/data/length"
        if actions:
            urimod = "/actions/length"
        return self.session.get(self.uri + urimod)

    def insert_array(self, datapoint_array):
        """given an array of datapoints, inserts them to the stream. This is different from append(),
        because it requires an array of valid datapoints, whereas append only requires the data portion
        of the datapoint, and fills out the rest::

            s.insert_array([{"d": 4, "t": time.time()},{"d": 5, "t": time.time()}])

        The optional `restamp` parameter specifies whether or not the database should rewrite the timestamps
        of datapoints which have a timestamp that is less than one that already exists in the database.

        That is, if restamp is False, and a datapoint has a timestamp less than a datapoint that already
        exists in the database, then the insert will fail. If restamp is True, then all datapoints
        with timestamps below the datapoints already in the database will have their timestamps overwritten
        to the same timestamp as the most recent datapoint hat already exists in the database, and the insert will
        succeed.
        """
        return self.session.post(self.uri + "/data", data=datapoint_array)

    def append(self, data):
        """inserts one datapoint with the given data, and appends it to
        the stream, using the current timestamp::

            s.append("Hello World!")

        """
        return self.insert_array([{"d": data, "t": time.time()}])

    def insert(self, data, timestamp=None):
        if timestamp is None:
            return self.append(data)
        return self.insert_array([{"t": timestamp, "d": data}])

    def remove(self, **kwargs):
        """
        Removes the given data from the stream
        """
        return self.session.delete(self.uri + "/data", params=kwargs)

    def __len__(self):
        return self.length()


registerObjectType("stream", Stream)
