[![Build Status](https://magnum.travis-ci.com/dkumor/connectordb.svg?token=wkfH9e4qB6qZhTstfRXR&branch=master)](https://magnum.travis-ci.com/dkumor/connectordb)

# ConnectorDB
An Open-Source database for Quantified Self and IoT. Please visit [the website](https://connectordb.github.io) for more information.

## Dependencies
You must have redis and postgres installed. ConnectorDB also requires at least golang 1.5, and needs 1.6 for http2.

To install the other necessary go dependencies, run:

```bash
make deps
```

## Building
The following will create all necessary binaries:

```bash
make
```

At that point, binaries are located in `bin`. Using the binaries in this folder, you can continue from the [setup tutorial](https://connectordb.github.io/download.html).

Note: On ubuntu your build might fail on npm step. This is because node is installed as nodejs.
`sudo ln -s /usr/bin/nodejs /usr/bin/node` should fix the issue.

## Testing
This will run all tests, spawning the necessary servers in the process (make sure you don't have any running connectordb instances):

```bash
make test
```

Note that this must be run _after_ build is completed.
