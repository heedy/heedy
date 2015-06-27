Benchmarking
=================

To start off, put your connectordb server into cpuprofile mode

```bash
connectordb -cpuprofile=cpu.pprof run database
```

Then spawn as many locust processes as your computer can handle (one slave per core)
```bash
locust -f <benchmark file> --master
locust -f <benchmark file> --slave
locust -f <benchmark file> --slave
```

Navigate to `localhost:8089` and then *POUND* the server with as many operations as it can handle.

Results
============

- writing stream was easily bottlenecked, with the majority of time spent
doing stuff to redis - I am curious as to what exactly was the issue
- gorilla is not a problem under real load - most of the time is spent doing our stuff.


```
benchmark      usernumber    median   average  max
Pinger         1000          10       16       325 (LOCUST LIMITED)
StreamWriter   1000          300      400      2300 (approximate)
```
