from .objects.timeseries import fixTimestamps, Timeseries, DatapointArray


def getQuery(ts, kwargs):
    fixTimestamps(kwargs)
    if isinstance(ts, Merge):
        kwargs["merge"] = ts.query
        return kwargs
    if isinstance(ts, Timeseries):
        ts = ts.id
    elif not isinstance(ts, str):
        raise Exception("Timeseries must be either ID string or a Timeseries object")
    kwargs["timeseries"] = ts
    return kwargs


class Merge:
    """Merge represents a query which allows to merge multiple timeseries into one
    when reading, with all the data merged together by increasing timestamp.
    The merge query is used as a constructor-type object::
        m = Merge(h)
        m.add("objectid",t1="now-10h")
        m.add(h.objects["objectid2"],t1="now-10h")
        result = m.run()
    """

    def __init__(self, h):
        """Given a heedy connection, begins the construction of a merge query"""
        self.h = h
        self.query = []

    def add(self, ts, **kwargs):
        self.query.append(getQuery(ts, kwargs))

    def run(self):
        return self.h.session.post(
            "/api/timeseries/dataset",
            {"data": {"merge": self.query}},
            f=lambda x: DatapointArray(x["data"]),
        )


class Dataset(object):
    """Heedy is capable of taking several separate unrelated timeseries, and based upon
    the chosen interpolation method, putting them all together to generate tabular data centered about
    either another timeseries' datapoints, or based upon time intervals.
    The underlying issue that Datasets solve is that in Heedy, timeseries are inherently unrelated.
    In most data stores, such as standard relational (SQL) databases, and even excel spreadsheets, data is in tabular
    form. That is, if we have measurements of temperature in our house and our mood, we have a table:
        +--------------+----------------------+
        | Mood Rating  | Room Temperature (F) |
        +==============+======================+
        | 7            | 73                   |
        +--------------+----------------------+
        | 3            | 84                   |
        +--------------+----------------------+
        | 5            | 79                   |
        +--------------+----------------------+
    The benefit of having such a table is that it is easy to perform data analysis. You know which temperature
    value corresponds to which mood rating. The downside of having such tables
    is that Mood Rating and Room Temperature must be directly related - a temperature measurement must be made
    each time a mood rating is given. Heedy has no such restrictions. Mood Rating and Room Temperature
    can be entirely separate sensors, which update data at their own rate. In Heedy, each timeseries
    can be inserted with any timestamp, and without regard for any other data.
    This separation of Timeseries makes data require some preprocessing and interpolation before it can be used
    for analysis. This is the purpose of the Dataset query. Heedy can put several streams together based
    upon chosen transforms and interpolators, returning a tabular structure which can readily be used for ML
    and statistical applications.
    There are two types of dataset queries
    :T-Dataset:
        T-Dataset: A dataset query which is generated based upon a time range. That is, you choose a time range and a
        time difference between elements of the dataset, and that is used to generate your dataset.
            +--------------+----------------------+
            | Timestamp    | Room Temperature (F) |
            +==============+======================+
            | 1pm          | 73                   |
            +--------------+----------------------+
            | 4pm          | 84                   |
            +--------------+----------------------+
            | 8pm          | 79                   |
            +--------------+----------------------+
        If I were to generate a T-dataset from 12pm to 8pm with dt=2 hours, using the interpolator "closest",
        I would get the following result:
            +--------------+----------------------+
            | Timestamp    | Room Temperature (F) |
            +==============+======================+
            | 12pm         | 73                   |
            +--------------+----------------------+
            | 2pm          | 73                   |
            +--------------+----------------------+
            | 4pm          | 84                   |
            +--------------+----------------------+
            | 6pm          | 84                   |
            +--------------+----------------------+
            | 8pm          | 79                   |
            +--------------+----------------------+
        The "closest" interpolator happens to return the datapoint closest to the given timestamp. There are many
        interpolators to choose from (described later).
        Hint: T-Datasets can be useful for plotting data (such as daily or weekly averages).
    :X-Dataset:
        X-datasets allow to generate datasets based not on evenly spaced timestamps, but based upon values of a timeseries.
        Suppose you have the following data:
            +-----------+--------------+---+-----------+----------------------+
            | Timestamp | Mood Rating  |   | Timestamp | Room Temperature (F) |
            +===========+==============+===+===========+======================+
            | 1pm       | 7            |   | 2pm       | 73                   |
            +-----------+--------------+---+-----------+----------------------+
            | 4pm       | 3            |   | 5pm       | 84                   |
            +-----------+--------------+---+-----------+----------------------+
            | 11pm      | 5            |   | 8pm       | 81                   |
            +-----------+--------------+---+-----------+----------------------+
            |           |              |   | 11pm      | 79                   |
            +-----------+--------------+---+-----------+----------------------+
        An X-dataset with X=Mood Rating, and the interpolator "closest" on Room Temperature would generate:
            +--------------+----------------------+
            | Mood Rating  | Room Temperature (F) |
            +==============+======================+
            | 7            | 73                   |
            +--------------+----------------------+
            | 3            | 84                   |
            +--------------+----------------------+
            | 5            | 79                   |
            +--------------+----------------------+
    :Interpolators:
        Interpolators are special functions which specify how exactly the data is supposed to be combined
        into a dataset. Any PipeScript script can be used as an interpolator, including "sum", "count" and other transforms.
        By default, the "closest" interpolator is used, which simply returns the datapoint closest to the reference timestamp.
    """

    def __init__(self, h, x=None, **kwargs):
        """In order to begin dataset generation, you need to specify the reference time range or stream.
        To generate a T-dataset::
            d = Dataset(h, t1=start, t2=end, dt=tchange)
        To generate an X-dataset::
            d = Dataset(h,"mytimeseriesid", t1=start, t2=end)
        Note that everywhere you insert a timeseries ID, you are also free to insert Timeseries objects
        or even Merge queries. The Dataset query in Heedy supports merges natively for each field.
        """
        self.h = h

        if x is not None:
            if "dt" in kwargs:
                raise Exception(
                    "Can't do both T-dataset and X-dataset at the same time"
                )
            self.query = getQuery(x, kwargs)
        elif "dt" in kwargs:
            fixTimestamps(kwargs)
            self.query = kwargs
        else:
            raise Exception("Dataset must have either x or dt parameter")

        self.query["dataset"] = {}

    def add(self, key, ts, **kwargs):
        """Adds the given timeseries to the query construction. Unless an interpolator is specified, "closest" will be used.
        You can insert a merge query instead of a timeseries::
            d = Dataset(h, t1="now-1h",t2="now",dt=10)
            d.add("temperature","timeseriesid")
            d.add("steps",h.objects["timeseriesid2"])
            m = Merge(h)
            m.add("id1")
            m.add("id2")
            d.add("altitude",m)
            result = d.run()
        """

        if key in self.query["dataset"]:
            raise Exception("The key already exists")

        self.query["dataset"][key] = getQuery(ts, kwargs)

    def run(self):
        """Runs the dataset query, and returns the result"""
        return self.h.session.post(
            "/api/timeseries/dataset",
            {"data": self.query},
            f=lambda x: DatapointArray(x["data"]),
        )
