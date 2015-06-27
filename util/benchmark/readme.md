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

Daniel - I need to find a good way to run locust, since it looks like locust is CPU-limited on my computer (I think my poor results are due to locust simply not being able to handle the requests I wanted to give it) - it took basically 100% of all my cpus.
The cool part is that while locust was tearing up my processor, the actual server was barely using any.

```
benchmark   usernumber    median   average  max
Pinger      1000          14       29       598
```

Joseph
```
benchmark   usernumber    median   average  max
Pinger      1000          ?        ?        ?
```
