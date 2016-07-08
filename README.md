[![Build Status](https://travis-ci.org/connectordb/connectordb.svg?branch=master)](https://travis-ci.org/connectordb/connectordb)
[![Gitter](https://badges.gitter.im/connectordb/connectordb.svg)](https://gitter.im/connectordb/connectordb?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

# ConnectorDB
A repository for all of your quantified-self data, and a control center for your IoT devices.

There already exist many apps and fitness trackers that gather and attempt to make sense of your data. Most of these services are isolated - your phone's fitness tracking software knows nothing about your browser's time-tracking extension. Furthermore, each app and service has its own method for downloading data (if they offer raw data at all!), which makes an all-encompassing analysis of life extremely tedious. ConnectorDB offers a self-hosted open-source alternative to these services. It allows every device you have to synchronize with one central database, which allows creating an in-depth picture of your life.

Please visit [the website](https://connectordb.github.io) for more information.

## Gather Data

This repository is just the backend server - to automatically gather data, you'll want to run the apps on your devices, which will automatically sync to ConnectorDB:

- [LaptopLogger](https://github.com/connectordb/connectordb-laptoplogger) - an app that gathers data about your computer usage (such as keypresses and active application)
- [Android App](https://github.com/connectordb/connectordb-android) - Gathers several metrics from your android phone in the background, including step count and current activity.
- [Python API](https://github.com/connectordb/connectordb-python) - gives full access to ConnectorDB, and allows you to write your own data-gathering apps (such as integrating custom sensors and devices), as well as data-analysis or IoT apps, which make use of your data (make your thermostat turn on in response to your phone's location).

## Rate Your life

ConnectorDB also has built-in support for manual data input - the [server's frontend](https://github.com/connectordb/connectordb-frontend) allows inputting any type of data, and has special support for star ratings - such as ratings of mood or productivity.

![Ratings](https://raw.githubusercontent.com/connectordb/connectordb/master/screenshot.png)

## Installing
Installation instructions, and precompiled binaries are [are available on the website](https://connectordb.github.io/download.html). You'll need a linux server, and if you want to use the android app, you will also need a domain name.

## Building

### Compile
You must have redis and postgres installed (and *not* running). ConnectorDB also requires at least golang 1.5, and needs 1.6 for http2.

To install the other necessary go dependencies, run:

```bash
make deps
```

To build ConnectorDB, you need to add the `src` directory to your gopath. After doing so, the following will create all necessary binaries:

```bash
make
```

At that point, binaries are located in `bin`. Using the binaries in this folder, you can continue from the [setup tutorial](https://connectordb.github.io/download.html).

Note: On ubuntu your build might fail on npm step. This is because node is installed as nodejs.
`sudo ln -s /usr/bin/nodejs /usr/bin/node` should fix the issue.

### Test
This will run all tests, spawning the necessary servers in the process (make sure you don't have any running connectordb instances):

```bash
make test
```

Note that this must be run _after_ build is completed.


### Windows

You will need to download the raw executables for gnatsd, redis, and postgres, and put them in the bin/dep directory after build. You'll also need to manually build the frontend, and put it in the bin/app directory, as well as manually copying site/www to bin.