# Apps

Heedy's apps are used as logical entities that interact with external services. 
In most cases, a plugin defines an app in its `heedy.conf`, which a user can then add to their account. The plugin then handles the app's features and synchronization with external data sources.

Apps can also be used directly without associated plugins if they have a access token defined.
An access token enables an external service/script to query the Heedy API with the permissions defined in the app's scope string. This allows an external service to sync with Heedy using an Oauth2 flow *(not yet implemented)*, or explicitly by asking the user to manually create an app with certain permissions and provide it the associated access token.

Once an auth token is obtained, logging into Heedy can be done by initializing the {class}`~heedy.App` object:

````{tab} Sync
```python
app = heedy.App("my_access_token",url="http://localhost:1324")
```
````

````{tab} Async
```python
app = heedy.App("my_access_token",url="http://localhost:1324",session="async")
```
````

## Listing & Accessing

While an app can be obtained directly from its access token, apps are also frequently accessed
and used by plugins. Apps can be listed given a set of constraints using the {attr}`~heedy.plugins.Plugin.apps` attribute of the {class}`~heedy.plugins.Plugin` object:

````{tab} Sync
```python
# Returns all enabled apps with the given plugin key (usually defined in heedy.conf)
apps = p.apps(plugin="myplugin:myapptype",enabled=True)
```
````

````{tab} Async
```python
# Returns all enabled apps with the given plugin key (usually defined in heedy.conf)
apps = await p.apps(plugin="myplugin:myapptype",enabled=True)
```
````

To list the apps beloning to a specific user, the constraint `owner="myuser"` can be used,
or the list can be generated from the {attr}`~heedy.User.apps` attribute of the {class}`~heedy.User` object, which automatically adds the owner constraint. 

If an app's ID is known, the corresponding {class}`~heedy.App` object can be retrieved with:

````{tab} Sync
```python
app = p.apps["appid123"]
print(app.read())
```
````

````{tab} Async
```python
app = p.apps["appid123"]
print(await app.read())
```
````

```{note}
Only plugins can list or read apps - when accessing heedy using an app's access token, the app can only access itself, and cannot see any other apps.
```

## Creating & Deleting

A Heedy plugin can create/delete apps, but an app logged in with an app token cannot. An app can be added with the {func}`~heedy.Apps.create` method:

````{tab} Sync
```python
app = p.apps.create("My App Name",
        owner="myusername"
        plugin=f"{p.name}:myapp"
    )
```
````

````{tab} Async
```python
app = await p.apps.create("My App Name",
        owner="myusername"
        plugin=f"{p.name}:myapp"
    )
```
````

The above is equivalent to:

````{tab} Sync
```python
app = p.users["myusername"].apps.create("My App Name",
        plugin=f"{p.name}:myapp"
    )
```
````

````{tab} Async
```python
app = await p.users["myusername"].apps.create("My App Name",
        plugin=f"{p.name}:myapp"
    )
```
````

Once the {class}`~heedy.App` object is retrieved, it can be deleted with the {func}`~heedy.App.delete` method:

````{tab} Sync
```python
app.delete()
```
````

````{tab} Async
```python
await app.delete()
```
````

## Reading & Updating Properties

Just like with {class}`~heedy.user` and {class}`~heedy.Object`, an {class}`~heedy.App`'s properties can be accessed directly as properties or as attributes:

````{tab} Sync
```python
# Does not use a server query - but it might say "self" before read() is called
# if logged in using the app token.
print(app.id) 
# Reads the app's display name from the server
print(app.name)
# Uses the previously read cached data, avoiding a server query
print(app["name"])
```
````

````{tab} Async
```python
# Does not use a server query - but it might say "self" before read() is called
# if logged in using the app token.
print(app.id) 
# Reads the app's display name from the server
print(await app.name)
# Uses the previously read cached data, avoiding a server query
print(app["name"])
```
````

To update the corresponding properties, the update function can be used,
or if in a sync session, the properties can also be directly set:

````{tab} Sync
```python
app.name = "My App" # Sets the app's display name
# The update function allows setting multiple properties in a single query
app.update(description="Syncs with a cool service")
```
````

````{tab} Async
```python
# The update function allows setting properties for the app:
await app.update(name="My App", description="Syncs with a cool service")

```
````

### Updating Access Token

An app access token cannot be directly modified. Instead, an update request is sent, 
and the server generates a new key itself. The newly generated key is returned as part of the
query result.

If accessing Heedy using an App token, and the token is updated, the session must be re-created with
the new token, since the old token is no longer valid for requests.

````{tab} Sync
```python
result = app.update(access_token=True)

# The old app object is no longer valid if it was logged in with the old access token
newapp = heedy.App(result["access_token"],url=app.session.url)
```
````

````{tab} Async
```python
result = await app.update(access_token=True)

# The old app object is no longer valid if it was logged in with the old access token
newapp = heedy.App(result["access_token"],url=app.session.url)
```
````

If the app has its access token set to `False` or `""`, the app will lose login capabilities,
and will only be accessible from plugins.

## API

(python_apps)=

### Apps

```{eval-rst}
.. autoclass:: heedy.Apps
    :members:
    :special-members:
    :inherited-members:
    :show-inheritance:
    :exclude-members: __init__, __weakref__
```

(python_app)=

### App

```{eval-rst}
.. autoclass:: heedy.App
    :members:
    :inherited-members:
    :show-inheritance:

    .. autoattribute:: props
```
