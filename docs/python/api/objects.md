# Objects

[Heedy Objects](heedy_objects) represent Heedy's core functionality. By default, Heedy supports
objects of the [`timeseries`](python_timeseries) type, but other types can be implemented by Heedy plugins, and can have
corresponding Python APIs registered with the Python client.

This page will describe how to use objects from the Python API, accessing them either from an app or a plugin.

## Listing & Accessing

Objects can be accessed using the [`objects`](python_objects) property in Apps, Plugins, or Users:

`````{tab} Sync
````{tab} App
```python
app = heedy.App("your_access_token","http://localhost:1324")
objs = app.objects() # list objects that belong to the app
print(objs)
```
````
````{tab} User
```python
app = heedy.App("your_access_token","http://localhost:1324")
myuser = app.owner # Gets the user that owns this app
objs = myuser.objects() # list all permitted objects belonging to the user
print(objs)
```
````
````{tab} Plugin
```python
p = heedy.Plugin(session="sync")
objs = p.objects() # access objects of any user on the server
print(objs)
```
````
`````

`````{tab} Async
````{tab} App
```python
app = heedy.App("your_access_token","http://localhost:1324")
objs = await app.objects() # list all objects that belong to the app
print(objs)
```
````
````{tab} User
```python
app = heedy.App("your_access_token","http://localhost:1324")
myuser = await app.owner # Gets the user that owns this app
objs = await myuser.objects() # list all objects that belong to the user
print(objs)
```
````
````{tab} Plugin
```python
p = heedy.Plugin(session="sync")
objs = await p.objects() # list all objects on the server
print(objs)
```
````
`````

The list returned from the `objects` function can be constrained, usually by object type, tags, or app key.
The tags are space-separated, and will match all objects with their superset.

````{tab} Sync
```python
app.objects(tags="steps fitbit",type="timeseries")
```
````

````{tab} Async
```python
await app.objects(tags="steps fitbit",type="timeseries")
```
````

```
[Timeseries{'access': 'read',
  'app': 'acf85b01-3c87-4813-9538-3c1120ecb2a3',
  'created_date': '2021-03-12',
  'description': '',
  'id': 'e1f4d4c2-d4f0-431a-bcaf-58c512fc7564',
  'key': 'steps',
  'meta': {'schema': {'type': 'number'}},
  'modified_date': '2021-09-27',
  'name': 'Steps',
  'owner': 'test',
  'owner_scope': 'read',
  'tags': 'fitbit steps charge5',
  'type': 'timeseries'}]
```

## Creating & Deleting

## Updating

## Timeseries Data

## Registering New Types

## API

(python_objects)=

### Objects

```{eval-rst}
.. autoclass:: heedy.Objects
    :members:
    :inherited-members:
    :show-inheritance:
```

(python_object)=

### Object

```{eval-rst}
.. autoclass:: heedy.Object
    :members:
    :inherited-members:
    :show-inheritance:
```

(python_objectmeta)=

#### Meta

```{eval-rst}
.. autoclass:: heedy.objects.objects.ObjectMeta
    :members:
    :inherited-members:
    :show-inheritance:
```

(python_timeseries)=

### Timeseries

```{eval-rst}
.. autoclass:: heedy.objects.timeseries.Timeseries
    :members:
    :inherited-members:
    :show-inheritance:
```

#### DatapointArray

```{eval-rst}
.. autoclass:: heedy.objects.timeseries.DatapointArray
    :members:
    :inherited-members:
    :show-inheritance:
```

(python_objectregistry)=

### Registry

```{eval-rst}
.. automodule:: heedy.objects.registry
    :members:
    :undoc-members:
    :inherited-members:
    :show-inheritance:
```
