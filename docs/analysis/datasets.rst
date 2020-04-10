Datasets
=============

Datasets are one of the most powerful features of Heedy. The underlying issue is simple: You have multiple timeseries of data
being gathered at the same time. Each series is independent, so they do not have synchronized timestamps.

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

While the independence of data streams is an extremely useful feature when it comes to gathering data, it makes putting streams together difficult.

The ideal format for analysis would be a single table - the equivalent of a spreadsheet, where you have
a temperature for each mood rating.

  +--------------+----------------------+
  | Mood Rating  | Room Temperature (F) |
  +==============+======================+
  | 7            | 73                   |
  +--------------+----------------------+
  | 3            | 84                   |
  +--------------+----------------------+
  | 5            | 79                   |
  +--------------+----------------------+

If you want to find if temperature affects your mood (or if your mood affects the temperature you set on your thermostat),
this table is much easier to work with than the two independent streams. This format can be directly plugged into the many machine learning
algorithms available, and is very easy to process and plot.


How to get there?
-----------------------

This is exactly the purpose of Datasets. A dataset is given a list of input data streams, and methods to use
when putting the streams together (called interpolators). It outputs a nice, tabular structure which can easily be used for analysis.

There are 2 types of dataset: T-datasets and X-datasets

X-dataset
---------------------------

An X-Dataset generates a dataset based upon a reference stream. This is the one that we would use to get the "desired" table shown above, given our sample data (first table). In particular, for the above example, we would set our reference stream to be mood, and use the interpolator `closest` on our temperature stream. This will get the closest temperature measurement to each mood rating. Note that we can add as many other streams as we want to this dataset, all of which will be interpolated to our mood measurements - the output would be one big table.

The `closest` interpolator (used for the temperature stream) happens to return the datapoint closest to the given timestamp. You can see a list of available interpolators [here](./interpolators.html).

T-dataset
---------------------------

A T-Dataset generates a dataset based upon timestamp. Suppose I have only one stream of data (although you can add as many as you want):

  +--------------+----------------------+
  | Timestamp    | Room Temperature (F) |
  +--------------+----------------------+
  | 1pm          | 73                   |
  +--------------+----------------------+
  | 4pm          | 84                   |
  +--------------+----------------------+
  | 8pm          | 79                   |
  +--------------+----------------------+

Now suppose I generate a T-Dataset from this data, from 12pm to 8pm, with an interval of dt=2 hours, using the interpolator `closest`. I would get the following result:


  +--------------+----------------------+
  | Timestamp    | Room Temperature (F) |
  +--------------+----------------------+
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


T-datasets are useful when you want to see how certain data changes over time, or want to plot multiple streams with same reference time.

Note that for datasets which do not include multiple streams of data, you can oftentimes get an equivalent effect using only data transforms.


