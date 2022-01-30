# Users

A heedy instance can have multiple [users](heedy_users), all of which have different apps, and their own objects. This page describes accessing users and their properties from the Heedy Python client.

## Listing & Accessing

There are two modes of access to users. The first, which allows listing and accessing arbitrary users in the Heedy instance is only available to plugins. A list of all users in the instance can be retrieved with the `plugin.users()` function from {class}`~heedy.Users`:

````{tab} Sync
```python
p = heedy.Plugin(session="sync")
p.users() # Returns a list of all users
```
````

````{tab} Async
```python
p = heedy.Plugin()
await p.users() # Returns a list of all users
```
````

To get a specific user by username, it is sufficient to access the {attr}`~heedy.plugins.Plugin.users` object as a dict:

````{tab} Sync
```python
p = heedy.Plugin(session="sync")
p.users["myuser"] # Returns user w/ username "myuser"
```
````

````{tab} Async
```python
p = heedy.Plugin()
p.users["myuser"] # Returns user w/ username "myuser"
```
````

On the other hand, when accessing from an app, the user that *owns* of the app can be accessed directly (if the app has the `owner:read` scope!):

````{tab} Sync
```python
app = heedy.App("your_access_token","http://localhost:1324")
app.owner # Returns the User object of the app's owner
```
````

````{tab} Async
```python
app = heedy.App("your_access_token","http://localhost:1324",session="async")
await app.owner # Returns the User object of the app's owner
```
````

Each {class}`~heedy.App` and {class}`~heedy.Object` element has a corresponding {attr}`~heedy.App.owner` property/attribute,
which can be read by both plugins and apps.

## Creating & Deleting

A Heedy plugin can create/delete users, but an app cannot. Users can be added using the {func}`~heedy.Users.create` method:

````{tab} Sync
```python
usr = p.users.create("myusername","mypassword")
```
````

````{tab} Async
```python
usr = await p.users.create("myusername","mypassword")
```
````

Likewise, once a {class}`~heedy.User` object is retrieved, the user can be deleted by calling its {func}`~heedy.User.delete` method:

````{tab} Sync
```python
usr.delete()
```
````

````{tab} Async
```python
await usr.delete()
```
````

## Reading & Updating Properties

Properties available on a user can be accessed directly as properties or attributes in the {class}`~heedy.User` object:

````{tab} Sync
```python
# Does not use a server query, since username is the user object's ID
print(usr.username) 
# Reads the user's display name from the server
print(usr.name)
# Uses the previously read cached data, avoiding a server query
print(usr["name"])
```
````

````{tab} Async
```python
# Does not use a server query, since username is the user object's ID
print(usr.username) 
# Reads the user's display name from the server
print(await usr.name)
# Uses the previously read cached data, avoiding a server query
print(usr["name"])
```
````

To update the corresponding properties, the update function can be used,
or if in a sync session, the properties can also be directly set:

````{tab} Sync
```python
usr.name = "Alan Turing" # Sets the user's display name
# The update function allows setting multiple properties in a single query
usr.update(description="I like computers")
```
````

````{tab} Async
```python
# The update function allows setting properties for the user:
await usr.update(name="Alan Turing", description="I like computers")

```
````

### Updating User Password

Since heedy plugins have full access to the Heedy database, they can also change a user's password.
Passwords in Heedy are hashed, so it is not possible to read an existing password, but it can be updated like a normal property:

````{tab} Sync
```python
usr.password= "newpassword"
```
````

````{tab} Async
```python
await usr.update(password="newpassword")
```
````

## API

(python_users)=

### Users

```{eval-rst}
.. autoclass:: heedy.Users
    :members:
    :special-members:
    :inherited-members:
    :show-inheritance:
    :exclude-members: __init__, __weakref__
```

(python_user)=

### User

```{eval-rst}
.. autoclass:: heedy.User
    :members:
    :inherited-members:
    :show-inheritance:

    .. autoattribute:: props
```
