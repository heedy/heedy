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
        if not t.startswith("now"):
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
        """Initializes the datapoint array. If given a filename, loads the data from the file"""
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
        d = list.__getitem__(self, key)
        if isinstance(key, slice):
            d = DatapointArray(d)
        return d

    def sort(self, f=lambda d: d["t"]):
        """Sort the data in-place by the given function. Uses the timestamp by default."""
        list.sort(self, key=f)
        return self

    def d(self):
        """
        Returns just the data portion of the datapoints as a list::

            DatapointArray([{"t": 12345, "d": "hi"}]).d() # ["hi"]
        """
        return list(map(lambda x: x["d"], self.raw()))

    def t(self):
        """
        Returns a list of just the timestamp portion of the datapoints.
        The timestamps are in python datetime's date format::

            DatapointArray([{"t": 12345, "d": "hi"}]).t() # [datetime.datetime(1969, 12, 31, 22, 25, 45)]
        """
        return list(map(lambda x: datetime.datetime.fromtimestamp(x["t"]), self.raw()))

    def dt(self):
        """
        Returns a list of just the durations of all datapoints::

            DatapointArray([
                {"t": 12345, "d": "hi", "dt": 10},
                {"t": 12346, "d": "hi"},
            ]).dt() # [10,0]

        """
        return list(map(lambda x: 0.0 if not "dt" in x else x["dt"], self.raw()))

    def merge(self, array):
        """
        Merges the current data with the given array.
        It assumes that the datapoints are formatted correctly for heedy, meaning
        that they are in the format::

            [{"t": unix timestamp, "d": data,"dt": duration (optional)}]

        The data does NOT need to be sorted by timestamp - this function sorts it for you
        """
        self.extend(array)
        self.sort()

        return self

    def raw(self):
        """
        Returns array as a raw python list. For cases where for some reason
        the :code:`DatapointArray` wrapper does not work for you.
        """
        return list.__getitem__(self, slice(None, None))

    def save(self, filename):
        """Writes the data to the given file::

            DatapointArray([{"t": unix timestamp, "d": data}]).save("myfile.json")

        The data can later be loaded using load.
        """
        with open(filename, "w") as f:
            json.dump(self, f)

    @staticmethod
    def load(filename):
        """
        Adds the data from a JSON file. The file is expected to be in datapoint format::

            d = DatapointArray.load("myfile.json")

        Can be used to read data dumped by :code:`save()`.
        """
        dpa = DatapointArray()
        with open(filename, "r") as f:
            dpa.merge(json.load(f))
        return dpa

    def tshift(self, t: float):
        """Shifts all timestamps in the datapoint array by the given number of seconds.
        It is the same as the 'tshift' transform.

        Warning: The shift is performed in-place! This means that it modifies the underlying array::

            d = DatapointArray([{"t":56,"d":1}])
            d.tshift(20)
            print(d) # [{"t":76,"d":1}]

        Args:
            t (float): Number of seconds to shift the timestamps by.
        Returns:
            self, the shifted datapoint array.
        """
        raw = self.raw()
        for i in range(len(raw)):
            raw[i]["t"] += t
        return self

    def sum(self):
        """
        Returns the sum of the data portions of all datapoints within::

            DatapointArray([
                {"t": 12345, "d": 1},
                {"t": 12346], "d": 3.5}
                ]).sum() # 4.5
        """
        raw = self.raw()
        s = 0
        for i in range(len(raw)):
            s += raw[i]["d"]
        return s

    def mean(self):
        """
        Gets the mean of the data portions of all datapoints within::

            DatapointArray([
                {"t": 12345, "d": 1},
                {"t": 12346], "d": 2}
                ]).mean() # 1.5

        """
        return self.sum() / float(len(self))

    def to_df(self):
        """Returns the data as a pandas dataframe. The dataframe has a "t" column that
        contains the timestamps as datetime objects. If the data has durations, there is a "dt" column as a timedelta.

        Finally, if the data is a number, string, or boolean, there is a "d" column that contains the data.
        Otherwise, if the data portion is an object, the data has a column for each key, separated by "_"::

            DatapointArray([
                {"t": 12345, "d": {"a": 1, "b": 2}},
                ]).to_df() # columns: t, d_a, d_b

        Returns:
            pandas.DataFrame: The data as a pandas dataframe.
        """
        import pandas

        df = pandas.json_normalize(self, sep="_")
        if len(df) > 0:
            df["t"] = pandas.to_datetime(df["t"], unit="s")
            if "dt" in df:
                df["dt"] = pandas.to_timedelta(df["dt"], unit="s")
        return df


class Timeseries(Object):

    output_type = "list"
    """
    A global property allowing you to specify the format in which timeseries data is returned by default. 
    Can be one of:

    - "list" (default): A :code:`DatapointArray` containing the data (i.e. a list of dicts), see :ref:`python_datapointarray`.
    - "dataframe": A pandas dataframe containing the data. This is the default used in the heedy notebook interface, and is equivalent to using :code:`to_df` on the :code:`DatapointArray` returned for the "list" type.

    """

    def __call__(self, **kwargs):
        """
        Returns the timeseries data matching the given constraints.

        .. tab:: Sync

            ::

                data = timeseries(t1="now-1h") # returns the last hour of data
                data = timeseries(t1="now-1h",output_type="dataframe") # returns the last hour of data, as a pandas dataframe
                data = timeseries(t1="jun 5",t2="jul 5",transform="sum") # returns the sum of data from jun 5th to jul 5th of this year.

        .. tab:: Async

            ::

                data = await timeseries(t1="now-1h") # returns the last hour of data
                data = await timeseries(t1="now-1h",output_type="dataframe") # returns the last hour of data, as a pandas dataframe
                data = await timeseries(t1="jun 5",t2="jul 5",transform="sum") # returns the sum of data from jun 5th to jul 5th of this year.

        Args:
            t1 (float or str): Only return datapoints with timestamp :code:`>=t1`. Can be a unix timestamp, a string such as "last month", "1pm" or "jun 5, 2019, 1pm" (any text supported by the `dateparser` library), or a relative time such as "now-1h" or "now-1d".
            t2 (float or str): Only return datapoints with timestamp :code:`<t2`, with identical semantics to t1.
            t (float or str): Return only the datapoint with the given exact timestamp (same semantics as t1)
            i1 (int): Only return datapoints with index :code:`>=i1` (negative numbers are relative to the end of the timeseries).
            i2 (int): Only return datapoints with index :code:`<i2`, with identical semantics to i1.
            i (int): Return only the datapoint with the given exact index (same semantics as i1).
            limit (int): The maximum number of datapoints to return.
            transform (str): The transform to apply to the data. See :ref:`pipescript`.
            output_type (str): The Python output type for this query. One of "list" or "dataframe". By default, :code:`Timeseries.output_type` is used.
        Returns:
            A :code:`DatapointArray` or a pandas dataframe, depending on the value of :code:`Timeseries.output_type` (or :code:`output_type` argument).
        Raises:
            HeedyException: If the server returns an error.
        """

        # Do we return a datapoint array or a pandas dataframe?
        conversion = lambda x: DatapointArray(x)
        if "output_type" in kwargs:
            if kwargs["output_type"] == "dataframe":
                conversion = lambda x: DatapointArray(x).to_df()
            del kwargs["output_type"]
        elif self.output_type == "dataframe":
            conversion = lambda x: DatapointArray(x).to_df()

        fixTimestamps(kwargs)
        return self.session.get(self.uri + "/timeseries", params=kwargs, f=conversion)

    def __getitem__(self, getrange):
        """Allows accessing the timeseries just as if it were just one big python array.

        .. tab:: Sync

            ::

                #Returns the most recent 5 datapoints from the timeseries
                ts[-5:]

                #Returns all the data the timeseries holds.
                ts[:]

                # Returns the most recent datapoint
                ts[-1]

        .. tab:: Async

            ::

                #Returns the most recent 5 datapoints from the timeseries
                await ts[-5:]

                #Returns all the data the timeseries holds.
                await ts[:]

                # Returns the most recent datapoint
                await ts[-1]


        This is equivalent to calling :code:`__call__` with `i1` and `i2` or `i` arguments.

        Note that if passed a string, it is equivalent to calling :code:`Object[]` with the string,
        meaning that it gets the cached value of the prop::

            assert ts["name"]==ts.cached_data["name"]

        Returns:
            A :code:`DatapointArray` or a pandas dataframe, depending on the value of :code:`Timeseries.output_type`,
            or if passed string, the corresponding cached property value (see :code:`Object.props`).
        Raises:
            HeedyException: If the server returns an error.
        """
        if isinstance(getrange, str):
            return super().__getitem__(getrange)
        if not isinstance(getrange, slice):
            # Return the single datapoint
            return self.session.f(self(i=getrange),lambda x: x[0])

        # The query is a slice - return the range
        qkwargs = {}
        if getrange.start is not None:
            qkwargs["i1"] = getrange.start
        if getrange.stop is not None:
            qkwargs["i2"] = getrange.stop
        return self(**qkwargs)

    def length(self):
        """Returns the number of datapoints in the timeseries.

        .. tab:: Sync

            ::

                len(timeseries) # equivalent to ts.length()

        .. tab:: Async

            ::

                await ts.length()

        Returns:
            The number of datapoints in the timeseries.
        Raises:
            HeedyException: If the server returns an error.

        """
        return self.session.get(self.uri + "/timeseries/length")

    def insert_array(self, datapoint_array, **kwargs):
        """
        Given an array of datapoints in the heedy format, insert them into the timeseries.

        .. tab:: Sync

            ::

                ts.insert_array([
                    {"d": 4, "t": time.time()},
                    {"d": 5, "t": time.time(), "dt": 5.3}
                    ])

        .. tab:: Async

            ::

                await ts.insert_array([
                    {"d": 4, "t": time.time()},
                    {"d": 5, "t": time.time(), "dt": 5.3}
                    ])



        Heedy's timeseries are indexed by timestamp, and datapoints in the series cannot have the same timestamp, or have overlapping durations.
        By default, heedy overwrites existing data with new data on insert when there is timestamp or duration overlap (`update` write method).
        When write method is set to `insert`, Heedy succeeds writing datapoints that don't overlap, but fails if it would affect existing data.
        Finally, if set to `append`, only appending is permitted to the timeseries, meaning attempting to write datapoints before the most recent one will fail.

        .. tab:: Sync

            ::

                ts.insert_array([{"d": 4, "t": 123456}]) # Timeseries has 4 at the timestamp
                ts.insert_array([{"d": 5, "t": 123456}]) # The 4 was replaced with a 5

                ts.insert_array([{"d": 6, "t": 123456}], method="update") # THROWS ERROR
                ts.insert_array([{"d": 7, "t": 123455}], method="update") # Succeeds (non-overlap)

                ts.insert_array([{"d": 8, "t": 123456}], method="append") # THROWS ERROR
                ts.insert_array([{"d": 9, "t": 123455}], method="append") # THROWS ERROR
                ts.insert_array([{"d": 10, "t": 123457}], method="append") # Succeeds (after existing)

        .. tab:: Async

            ::

                await ts.insert_array([{"d": 4, "t": 123456}]) # Timeseries has 4 at the timestamp
                await ts.insert_array([{"d": 5, "t": 123456}]) # The 4 was replaced with a 5

                await ts.insert_array([{"d": 6, "t": 123456}], method="update") # THROWS ERROR
                await ts.insert_array([{"d": 7, "t": 123455}], method="update") # Succeeds (non-overlap)

                await ts.insert_array([{"d": 8, "t": 123456}], method="append") # THROWS ERROR
                await ts.insert_array([{"d": 9, "t": 123455}], method="append") # THROWS ERROR
                await ts.insert_array([{"d": 10, "t": 123457}], method="append") # Succeeds (after existing)

        Args:
            datapoint_array (list): A list of dicts, with each dictionary having the following keys:

                * "d" (json-convertible): The datapoint value.
                * "t" (float): The timestamp of the datapoint, in unix seconds.
                * "dt" (float,optional): The duration of the datapoint, in seconds.

            method (str, optional): The method to use when inserting datapoints. One of:

                * "update"
                    Insert datapoints, overwriting existing ones if they have the same timestamp or overlap.
                * "insert"
                    Insert datapoints, throwing an error if a timestamp conflicts with an existing one.
                * "append"
                    Only permit appends, meaning that no timestamp in the inserted array is <= any existing timestamp,
                    and is < and existing timestamp+duration.
        Raises:
            HeedyException: If the server returns an error.
        """
        return self.session.post(
            self.uri + "/timeseries", data=datapoint_array, params=kwargs
        )

    def append(self, data, duration=0):
        """Shorthand insert function, inserts the given data into the timeseries with the current timestamp.

        .. tab:: Sync

            ::

                ts.append("Hello World!")

        .. tab:: Async

            ::

                await ts.append("Hello World!")

        Equivalent to calling::

            ts.insert(data,duration=duration)

        Args:
            data (json-convertible): The value to insert
            duration (float, optional): The duration of the datapoint, in seconds.
        Raises:
            HeedyException: If the server returns an error.

        """
        return self.insert_array([{"d": data, "t": time.time(), "dt": duration}])

    def insert(self, data, timestamp=None, duration=0):
        """
        Inserts the given data into the timeseries at the given timestamp.

        .. tab:: Sync

            ::

                ts.insert("Hello World!")

        .. tab:: Async

            ::

                await ts.insert("Hello World!")

        Equivalent to calling::

            ts.insert_array([{"d": data, "t": timestamp,"dt":duration}])

        Args:
            data (json-convertible): The value to insert
            timestamp (float, optional): The timestamp of the datapoint, in unix seconds. If none given, current time is used.
            duration (float, optional): The duration of the datapoint, in seconds.
        Raises:
            HeedyException: If the server returns an error.

        """
        if timestamp is None:
            return self.append(data, duration)
        return self.insert_array([{"t": timestamp, "d": data, "dt": duration}])

    def remove(self, **kwargs):
        """
        Removes all datapoints satisfying the given constraints.

        .. tab:: Sync

            ::

                ts.remove(t1="now-1h") # remove the last hour of data
                ts.remove(i=-1) # removes the most recent datapoint

        .. tab:: Async

            ::

                await ts.remove(t1="now-1h") # remove the last hour of data
                await ts.remove(i=-1) # removes the most recent datapoint

        Args:
            t1 (float or str): Only remove datapoints with timestamp :code:`>=t1`. Can be a unix timestamp, a string such as "last month", "1pm" or "jun 5, 2019, 1pm" (any text supported by the `dateparser` library), or a relative time such as "now-1h" or "now-1d".
            t2 (float or str): Only remove datapoints with timestamp :code:`<t2`, with identical semantics to t1.
            t (float or str): Remove only the datapoint with the given exact timestamp (same semantics as t1)
            i1 (int): Only remove datapoints with index :code:`>=i1` (negative numbers are relative to the end of the timeseries).
            i2 (int): Only remove datapoints with index :code:`<i2`, with identical semantics to i1.
            i (int): Remove only the datapoint with the given exact index (same semantics as i1).
        Raises:
            HeedyException: If the server returns an error.
        """
        return self.session.delete(self.uri + "/timeseries", params=kwargs)

    def save(self, filename):
        """Saves the entire timeseries data as JSON to the given filename:

        .. tab:: Sync

            ::

                ts.save("myts.json")

        .. tab:: Async

            ::

                await ts.save("myts.json")

        The data can then be loaded using :code:`DatapointArray.load()`.

        Raises:
            HeedyException: If the server returns an error.

        """
        return self.session.get(
            self.uri + "/timeseries", f=lambda x: DatapointArray(x).save(filename)
        )

    def load(self, filename, **kwargs):
        """Loads timeseries data JSON from the given file to the timeseries:

        .. tab:: Sync

            ::

                ts.load("myts.json")

        .. tab:: Async

            ::

                await ts.load("myts.json")

        This function can load data saved with the :code:`save` function.

        Raises:
            HeedyException: If the server returns an error.

        """
        return self.insert_array(DatapointArray.load(filename), **kwargs)

    def __len__(self):
        return self.length()


registerObjectType("timeseries", Timeseries)
