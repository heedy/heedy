[![Build Status](https://magnum.travis-ci.com/dkumor/connectordb.svg?token=wkfH9e4qB6qZhTstfRXR&branch=master)](https://magnum.travis-ci.com/dkumor/connectordb)

# ConnectorDB
An Open-Source database for Quantified Self and IoT. Please visit [the website](https://connectordb.github.io) for more information.

## Dependencies
You must have redis and postgres installed. ConnectorDB also requires at least golang 1.5, and needs 1.6 for http2.

To install the other necessary go dependencies, run:

```bash
make go-dependencies
make submodules
```

## Building
The following will create all necessary binaries, and download the default web interface:

```bash
make
```

At that point, binaries are located in `bin`. Using the binaries in this folder, you can continue from the [setup tutorial](https://connectordb.github.io/download.html).

## Testing
This will run all tests, spawning the necessary servers in the process (make sure you don't have any running connectordb instances):

```bash
make test
```

Note that this must be run _after_ build is completed.
