# The asset manager handles the ConnectorDB assets.

Assets have the following directory structure:

/plugins - folder holding all plugins
./{plugin-name}
./assets - a folder mirroring the asset manager structure containing file replacements
plugin.conf - plugin configuration file
./src - a folder holding python source code
README.md - A readme file that holds the plugin description
/www - The website to show non-logged-in users
index.html (golang template)
login.html (golang template - login page to show whenever login is required. Can be shown at other URLs)
/app - the app to show logged-in users
index.html (golang template - the core html that is served for ALL pages)
/setup - the website to show in setup mode
index.html (golang template - served for ALL pages)
/database - Empty when starting, in this folder the database is held
connectordb.conf - The ConnectorDB configuration file. It is special, in that it is not replaced wholesale by plugins,
but rather the options defined in each plugin are overlayed together into a full file.

The root directory of the database is the "final" overlay, in that it overwrites all plugins.

# The Router gives the following format

/ <- loads /www/index.html for non-logged-in users, and /app/index.html for logged-in users
/www/ - hosts the www folder
/app/ - hosts the app folder
/api/2/ - hosts the ConnectorDB API
/api/2/p/{plugin-name}/ - hosts the API for the given plugin
/api/2/c/ - hosts the ConnectorDB built-in Core API

# The asset manager is given:

- A list of active plugins
- The path to the directory where cdb was initialized

It returns:

1. The full configuration based on active plugins
2. An overlay http.Filesystem for /www, /app, /setup that stays updated with plugin changes
