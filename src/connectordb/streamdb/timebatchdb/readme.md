TimebatchDB
===================

A very simple key-{time series} store.

A time series is first inserted into a redis instance, since datapoints can come in in small numbers. Then, once the number of datapoints in the time series reaches
a given batch size, a background process transparently encodes them into a compressed batch and stores in the preferred database (as of writing postgres).

This framework enables extremely fast inserts (thanks redis!), as well as storage of enormous amounts of compressed data
(each time series is stored long term in compressed chunks of hundreds of data points).

This is all done transparently to the user - all you see is a database of time series accessible by key, data ranges of which are readily queryable by timestamp or by index.

Timebatchdb makes no assumptions about the character of data stored - data is just arbitrary byte arrays.
