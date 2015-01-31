[![Build Status](https://magnum.travis-ci.com/dkumor/connectordb.svg?token=wkfH9e4qB6qZhTstfRXR&branch=master)](https://magnum.travis-ci.com/dkumor/connectordb)

ConnectorDB
=========================
A database that connects stuff

## Dependencies
You must have redis installed. To install the other necessary go dependencies, run:
```bash
make dependencies
```

## Building

The following will create all necessary binaries
```bash
make
```

At that point, binaries are located in `/bin`. Good default config files are located in `/config`.

## Testing

This will run all tests, spawning the necessary servers in the process:
```bash
make test
```

#### Manual Testing:
Manually running tests is done by running, all in different terminals:
```bash
redis-server config/redis.conf
```
```bash
./bin/gnatsd -c config/gnatsd.conf
```

Then:
```bash
go test streamdb/...
```
This allows to keep redis and gnatsd running while coding.
