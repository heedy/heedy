# connectordb_web

The web component of the ConnectorDB database. This is the underlying code of the default application that ships with ConnectorDB.

## Installing

Debugging the code here requires installing the files in this folder (cloning the repository) into `site/app` in the ConnectorDB code. This should be done automatically when cloning ConnectorDB.

In order to download the dependencies necessary to run the app, run `bower update`.

To modify the code (and in order to have an easy method for debugging), once connectordb is built, go to the `bin/app` directory, delete all files within, and clone this repository into it. This allows you to modify the site as ConnectorDB is running.
