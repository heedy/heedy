[![Build Status](https://magnum.travis-ci.com/dkumor/connectordb.svg?token=wkfH9e4qB6qZhTstfRXR&branch=master)](https://magnum.travis-ci.com/dkumor/connectordb)

ConnectorDB
=========================
A database that connects stuff

## Dependencies
You must have redis and postgres installed. To install the other necessary go dependencies, run:
```bash
make dependencies
```

## Building

The following will create all necessary binaries
```bash
make
```

At that point, binaries are located in `/bin`.

To have the python libs work, go to `src/clients/python` and follow instructions in readme there.

## Testing

This will run all tests, spawning the necessary servers in the process:
```bash
make test
```
