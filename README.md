<a href="https://connectordb.io"><img src="https://raw.githubusercontent.com/connectordb/branding/master/title_logo_dark.png" width="500"/></a>


[![Build Status](https://img.shields.io/travis/connectordb/connectordb.svg?style=flat-square&label=linux%2fdarwin+build)](https://travis-ci.org/connectordb/connectordb)
[![AppVeyor](https://img.shields.io/appveyor/ci/dkumor/connectordb.svg?style=flat-square&label=windows+build)](https://ci.appveyor.com/project/dkumor/connectordb)
[![Gitter](https://img.shields.io/gitter/room/connectordb/connectordb.svg?maxAge=2592000&style=flat-square)](https://gitter.im/connectordb/connectordb?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

A repository for all of your quantified-self data, and a control center for your IoT devices.

There already exist many apps and fitness trackers that gather and attempt to make sense of your data. Most of these services are isolated - your phone's fitness tracking software knows nothing about your browser's time-tracking extension. Furthermore, each app and service has its own method for downloading data (if they offer raw data at all!), which makes an all-encompassing analysis of life extremely tedious. ConnectorDB offers a self-hosted open-source alternative to these services. It allows every device you have to synchronize with one central database, which allows creating an in-depth picture of your life.

Please visit [the website](https://connectordb.io) for more information.

## Gather Data

This repository is just the backend server - to automatically gather data, you'll want to run the apps on your devices, which will automatically sync to ConnectorDB:

- [LaptopLogger](https://github.com/connectordb/connectordb-laptoplogger) - an app that gathers data about your computer usage (such as keypresses and active application). It is included by default in the windows desktop version of ConnectorDB.
- [Android App](https://github.com/connectordb/connectordb-android) - Gathers several metrics from your android phone in the background, including step count and current activity.
- [Chrome Extension](https://github.com/connectordb/connectordb-chrome) - Gathers your web browsing history, allowing you to keep track of how much time you spend on various websites.
- [Python API](https://github.com/connectordb/connectordb-python) - gives full access to ConnectorDB, and allows you to write your own data-gathering apps (such as integrating custom sensors and devices), as well as data-analysis or IoT apps, which make use of your data (make your thermostat turn on in response to your phone's location).

## Rate Your life

ConnectorDB also has built-in support for manual data input - the [server's frontend](https://github.com/connectordb/connectordb-frontend) allows inputting any type of data, and has special support for star ratings - such as ratings of mood or productivity.

![Ratings](https://raw.githubusercontent.com/connectordb/connectordb/master/screenshot.png)

## Installing
Installation instructions, and precompiled binaries are [are available on the website](https://connectordb.io/download/).

Development builds are available [here](https://keybase.pub/dkumor/connectordb). These are usually direct builds of master, and as such might be less stable.

## Building

To perform a full build of ConnectorDB you will need a linux machine, preferably arch or a recent version of ubuntu, and corresponding cross-compilers.

While technically builds can be performed on windows/OSX/Raspberry Pi, they are not officially supported. OSX/Raspberry Pi builds should "just work", with the following caveats:

- Raspbian has an old version of Redis, so you will need to download the source of Redis >3.0 and compile, putting redis-server and redis-cli in `bin/dep` folder. Look at the arm portion of `makerelease` for specific instructions.
- Windows has issues setting up NPM. If building on Windows, it is recommended that you use a precompiled version of the frontend, put in `bin/app`.



### Compile
You must have redis and postgres installed (and *not* running). ConnectorDB also requires golang 1.7.

To install the other necessary go dependencies, run:

```bash
make deps
```

To build ConnectorDB, you need to add the `connectordb` directory to your gopath. After doing so, the following will create all necessary binaries:

```bash
make
```

At that point, binaries are located in `bin`. Using the binaries in this folder, you can continue from the [setup tutorial](https://connectordb.io/download.html).

Note: On ubuntu your build might fail on npm step. This is because node is installed as nodejs.
`sudo ln -s /usr/bin/nodejs /usr/bin/node` should fix the issue.

### Test
This will run all tests, spawning the necessary servers in the process (make sure you don't have any running connectordb instances):

```bash
make test
```

Note that this must be run _after_ build is completed.
