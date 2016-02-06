<<<<<<< HEAD
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
=======
# connectordb_web

The web component of the ConnectorDB database. This is the underlying code of the default application that ships with ConnectorDB.

## Installing

Debugging the code here requires installing the files in this folder (cloning the repository) into `site/app` in the ConnectorDB code. This should be done automatically when cloning ConnectorDB.

In order to download the dependencies necessary to run the app, run `bower update`.

To modify the code (and in order to have an easy method for debugging), once connectordb is built, go to the `bin/app` directory, delete all files within, and clone this repository into it. This allows you to modify the site as ConnectorDB is running.
>>>>>>> c10eaf69048898a8f592217eca771431c2e18d57
