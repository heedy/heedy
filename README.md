[![Build Status](https://magnum.travis-ci.com/dkumor/connectordb.svg?token=wkfH9e4qB6qZhTstfRXR&branch=master)](https://magnum.travis-ci.com/dkumor/connectordb)

ConnectorDB
=========================
A database that connects stuff


## Building

The dependencies include `redis`, and all the go packages included in the `install` section of `.travis.yml`.

The `/tools` directory contains standalone executables, each of which needs to be individually compiled with a command such as:

```bash
go build ./tools/dbwriter.go
```

Lastly, gnatsd needs to be compiled from source:
```bash
go build github.com/apcera/gnatsd
```

## Running

The servers in the `before_script` part of `.travis.yml` need to be running in the background before the rest can run.
