[![Build Status](https://travis-ci.org/connectordb/connectordb-frontend.svg?branch=master)](https://travis-ci.org/connectordb/connectordb-frontend)

# ConnectorDB Frontend App

This is the web app used by default in ConnectorDB. It contains all of the relevant analysis, display and plotting code.


## Building

To debug the app, you will need a functioning ConnectorDB development environment.

To start off, build the ConnectorDB database:

```bash
git clone https://github.com/connectordb/connectordb
cd connectordb
make deps
make
```

After build completes, start the ConnectorDB server:

```bash
./bin/connectordb create testdb
./bin/connectordb start testdb --join
```

Open `localhost:3124`, where you will see the login screen - create a new user.

Note: On ubuntu your build might fail on npm step. This is because node is installed as nodejs.
`sudo ln -s /usr/bin/nodejs /usr/bin/node` should fix the issue.

### Setting up webapp development

Once you have the server running, you will want to start on the app itself. Web portions of ConnectorDB (including alternate web frontends) should be developed from the `site` subdirectory.

This app is pulled by default during `make deps`. To begin development, we update to the newest version (which might be newer than the one used by ConnectorDB):

```
cd site/app
git checkout master
git pull
```

Now you can start live updates to the code, which will automatically be reflected in your ConnectorDB server:

```
npm run dev
```

NOTE: npm scripts must be run in the `site/app` directory to work correctly.


### Tools

The frontend app is built with react-redux. In order to help with debugging, you should download the [Redux DevTools Extension](https://github.com/zalmoxisus/redux-devtools-extension), and add `?debug_session=1` to the end of your URL, so that you can persist the state between app code modifications.


![ratings](https://raw.githubusercontent.com/connectordb/connectordb-frontend/master/screenshots/ratings.png)
![ratings](https://raw.githubusercontent.com/connectordb/connectordb-frontend/master/screenshots/android_data.png)
![ratings](https://raw.githubusercontent.com/connectordb/connectordb-frontend/master/screenshots/data.png)
