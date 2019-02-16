
Upgrading ConnectorDB
=======================================

If using an old version of ConnectorDB, you will probably want to upgrade to the latest release.
Since the database format has changed, you will need to export all of your data from ConnectorDB,
and reimport it into your new server. You will, unfortunately, not be able to use the old database,
as in-place upgrades from alpha releases are not supported.


## Exporting Data

### From Local Server

Newer versions of ConnectorDB offer direct built-in support for exporting and importing data.
You simply need to start the backend servers, and run an export. If you already have a ConnectorDB
server running, you can use just the export command.

```
connectordb start mydatabase --backend
connectordb export mydatabase ./exportfolder
connectordb stop mydatabase
```

Unfortunately, older versions of ConnectorDB (particularly version 0.3.0a1) did not have this feature.
You can still export your data from these versions by following the Remote export instructions (below).

### From Remote Server

You do not need to have direct access to the server in order to export all of your data. The ConnectorDB
Python API offers a convenient command which can generate an export of your data directly:

```
pip install connectordb
```

```python
import connectordb

cdb = connectordb.ConnectorDB("username","password","https://myserver")
cdb.user.export("exportfolder")
```

The above python code will [export all data and API keys for your user](http://connectordb-python.readthedocs.io/en/latest/connectordb.html#connectordb._user.User.export). The export format can be directly
read by ConnectorDB's built-in import command. Note however that the exported user will not have a password. This is because the REST API does not permit accessing password hashes. When imported, the user's password will automatically be set to the username.

The Python API also has commands for exporting [devices](http://connectordb-python.readthedocs.io/en/latest/connectordb.html#connectordb._user.User.export), or even [single streams](http://connectordb-python.readthedocs.io/en/latest/connectordb.html#connectordb._stream.Stream.export).

## Importing Data

ConnectorDB supports importing data directly, and through the python API. With the Python API, you have fine-grained control over your import. However,
only a local server import can preserve user passwords.

### To Local Server

Importing to a local server is the only way to preserve passwords if the export was generated through `connectordb` itself. These exports contain hashed passwords,
which can be directly imported by the `connectordb import` command:

```
connectordb start mydatabase --backend
connectordb import mydatabase ./exportfolder
connectordb stop mydatabase
```

### To Remote Server

You can import exported data directly through the Python API:

```
import connectordb

cdb = connectordb.ConnectorDB("username","password","https://myserver")
cdb.import_users("exportfolder")
```

Make sure that your ConnectorDB user has the permissions necessary to create a new user! Also note that the new user's password will be the username. You should make sure
to change the password immediately after importing!

If you have an export, but would just like to import certain devices
to your current user, you can use


```
cdb.user.import_device("exportfolder/myuser/mydevice")
```
