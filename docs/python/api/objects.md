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
app = heedy.App("your_access_token","http://localhost:1324",session="async")
objs = await app.objects() # list all objects that belong to the app
print(objs)
```
````
````{tab} User
```python
app = heedy.App("your_access_token","http://localhost:1324",session="async")
myuser = await app.owner # Gets the user that owns this app
objs = await myuser.objects() # list all objects that belong to the user
print(objs)
```
````
````{tab} Plugin
```python
p = heedy.Plugin()
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

One can create new objects of a given type by calling the `create` method in [`objects`](python_objects):

````{tab} Sync
```python
obj = app.objects.create("My Timeseries",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata",
                    key="myts")
```
````

````{tab} Async
```python
obj = await app.objects.create("My Timeseries",
                    type="timeseries",
                    meta={"schema":{"type":"number"}},
                    tags="myts mydata",
                    key="myts")
```
````

Only the first argument, the object name, is required. If not explicitly specified, the object type will be `timeseries`, and all other fields will be empty.

When creating an object for an app, it is useful to give it a `key`.
Keys are unique per-app, meaning that the app can have only one object with the given key. The object then can be retrieved using:

````{tab} Sync
```python
obj = app.objects(key="myts")[0]
```
````

````{tab} Async
```python
obj = (await app.objects(key="myts"))[0]
```
````

Finally, to delete an object and all of its data, one can use the delete method:

````{tab} Sync
```python
obj.delete()
```
````

````{tab} Async
```python
await obj.delete()
```
````

```{note}
Plugins have administrative access to the database, but to create an object belonging to the app when logged in using an App token, the app must have `self.objects:create` scope. Similarly, deleting an object requires the `self.objects:delete` scope. Finally, editing the object's properties requires the `self.objects:update` scope. Writing an object's content/data requires the `self.objects:write` scope. To give an app full access to manage its own objects, you can give it `self.objects` super-scope.
```

## Reading & Updating Properties

All properties available on the object can be accessed directly as properties in Python in two ways:

````{tab} Sync
```python
# Reads the key property from the server
assert obj.key == "myts"
# Uses the previously read cached data, avoiding a server query
assert obj["key"] == "myts"
```
````

````{tab} Async
```python
# Reads the key property from the server
assert (await obj.key) == "myts"
# Uses the previously read cached data, avoiding a server query
assert obj["key"] == "myts"
```
````

When trying to access cached properties for an object that was not yet read from the server, `obj["key"]` will return a `KeyError`. To refresh the cache, it is sufficient to run:

````{tab} Sync
```python
obj.read()
assert obj["key"]=="myts"
```
````

````{tab} Async
```python
await obj.read()
assert obj["key"]=="myts"
```
````

The `read` function returns a dict containing the props. Any read or update for the object will also refresh the cache with new data.

To update properties, the update method can be used, or the props can directly be set if in a `sync` session:

````{tab} Sync
```python
# Both of these lines do the same thing.
obj.description = "My description"
obj.update(description="My description")
```
````

````{tab} Async
```python
#
#
await obj.update(description="My description")
```
````

(python_object_metadata_info)=

### Updating Metadata

Each object has a `meta` property, which is defined separately for every object type. For example, the `timeseries` object type defines `meta` to contain a JSON schema describing the data type of each datapoint in the series:

```
[Timeseries{...
  'meta': {'schema': {'type': 'number'}},
  ...
  'type': 'timeseries'}]
```

Since the `meta` property is a dict that can contain multiple sub-properties,
it has special handling in the API. For example, updating an object to set its meta property will _merge_ the new values with the old ones:

````{tab} Sync
```python
obj.meta() # Read meta from server ={"a": 1,"b":2,"c":3}
obj.meta = {"c": 4,"b": None} # Update the b,c values
print(obj.meta) # {"a": 1,"c":4} # The update results are cached
```
````

````{tab} Async
```python
await obj.meta() # Read meta from server ={"a": 1,"b":2,"c":3}
await obj.update(meta={"c": 4,"b": None}) # Update the b,c values
print(obj.meta) # {"a": 1,"c":4} # The update results are cached
```
````

Setting a key in `meta` to `None` resets it to its default value, so deleting a timeseries schema will simply reset the schema to `{}` rather than removing the key.

Finally, for simplicity, the Python API supports direct modification of the meta object:

````{tab} Sync
```python
obj.meta.schema = {"type":"string"}
del obj.meta.b
```
````

````{tab} Async
```python
await obj.meta.update(schema={"type":"string"})
await obj.meta.delete("b")
```
````

## Timeseries Data

The `timeseries` object type is built into heedy by default. A timeseries can be considered an array of datapoints of the form:

```javascript
{
    "t": 1234556789.0, // unix timestamp in seconds,
    "d": 2, // data associated with the timestamp,
    "dt": 60, //optional duration of the datapoint in seconds, [t,t+dt)
}
```

One can optionally specify a JSON schema as the `schema` property of object metadata, which then constrains the data portion of future datapoints to the schema (see example in [the metadata section above](python_object_metadata_info)).

### Reading Data

Suppose we have an object of type `timeseries`.
Just like in Python arrays, heedy timeseries can be accessed by index, with negative numbers allowing indexing from the end of the series. One can also directly query data ranges and series length:

````{tab} Sync
```python
len(obj)    # Number of datapoints in the series
obj[2:5]    # Query by index range
obj[-1]     # Get most recent datapoint
```
````

````{tab} Async
```python
await obj.length()      # Number of datapoints in the series
await obj(i1=2,i2=5)    # Query by index range
await obj(i=-1)[0]      # Get most recent datapoint
```
````

The timeseries data can also be accessed by timestamp and time range. Time ranges can be of two types: relative and absolute. A relative time range uses a string that specifies a timestamp relative to the current time (s,m,h,d,w,mo,y):

````{tab} Sync
```python
obj(t1="now-1w") # Returns the past week of data
```
````

````{tab} Async
```python
await obj(t1="now-1w") # Returns the past week of data
```
````

The timeseries can also directly be accessed using the unix timestamp (in seconds) as the time range specifier:

````{tab} Sync
```python
# Returns the data from 2 hours ago to 1 hour ago
obj(t1=time.time()-2*60*60,t2=time.time()-60*60)
```
````

````{tab} Async
```python
# Returns the data from 2 hours ago to 1 hour ago
await obj(t1=time.time()-2*60*60,t2=time.time()-60*60)
```
````

#### PipeScript Transforms

When querying data from a timeseries, you can also specify a [server-side transform](pipescript) of the data, allowing aggregation and direct processing:

````{tab} Sync
```python
# Returns the sum of data for the past week
obj(t1="now-1w",transform="sum")
```
````

````{tab} Async
```python
# Returns the sum of data for the past week
await obj(t1="now-1w",transform="sum")
```
````

#### Output DatapointArray

When reading timeseries data, by default the result is returned as a subclass of `list` with a couple useful add-ons:

````{tab} Sync
```python
data = obj(t1="now-1w")

# Returns an array of just data portions of the result
data.d()

# Returns a pandas DataFrame of the timeseries data
data.to_df()

# Writes the data to a file
data.write("myfile.json")
```
````

````{tab} Async
```python
data = await obj(t1="now-1w")

# Returns an array of just data portions of the result
data.d()

# Returns a pandas DataFrame of the timeseries data
data.to_df()

# Writes the data to a file
data.save("myfile.json")
```
````

Timeseries objects in Python can be configured to return `pandas.DataFrames` directly instead (as is done in the heedy notebook plugin), or per query:

````{tab} Sync
```python
from heedy import Timeseries
Timeseries.output_type="dataframe" # Return pandas.DataFrames by default

obj(i1=-10,output_type="list") # override the global configuration for this query
```
````

````{tab} Async
```python
from heedy import Timeseries
Timeseries.output_type="dataframe" # Return pandas.DataFrame by default

await obj(i1=-10,output_type="list") # override the global configuration for this query
```
````

### Writing Data

If we have write access (`write` scope), we can append a new datapoint to the timeseries directly, or write an array of data at once:

````{tab} Sync
```python
# Add a datapoint 5 with current timestamp to the series
obj.append(5)
assert obj[-1]["d"]==5

# Insert the given array of data
obj.insert_array([{"d": 6, "t": time.time()},{"d": 7, "t": time.time(), "dt": 5.3}])
```
````

````{tab} Async
```python
# Add a datapoint 5 with current timestamp to the series
await obj.append(5)
assert (await obj(i=-1))["d"]==5

# Insert the given array of data
await obj.insert_array([{"d": 6, "t": time.time()},{"d": 7, "t": time.time(), "dt": 5.3}])
```
````

### Removing Data

Removing data from a timeseries has identical semantics to querying data. In other words, it is sufficient to specify the range:

````{tab} Sync
```python
# Remove the last month of data
obj.remove(t1="now-1mo")

# Timeseries are indexed by timestamp, so a specific datapoint can be removed
# by calling remove with its timestamp
dp = obj[-1]
obj.remove(t=dp["t"])

# This is equivalent to the above:
obj.remove(i=-1)
```
````

````{tab} Async
```python
# Remove the last month of data
await obj.remove(t1="now-1mo")

# Timeseries are indexed by timestamp, so a specific datapoint can be removed
# by calling remove with its timestamp
dp = await obj[-1]
await obj.remove(t=dp["t"])

# This is equivalent to the above:
await obj.remove(i=-1)
```
````

## API

(python_objects)=

### Objects

```{eval-rst}
.. autoclass:: heedy.Objects
    :members:
    :special-members:
    :show-inheritance:
    :exclude-members: __init__
```

(python_object)=

### Object

```{eval-rst}
.. autoclass:: heedy.Object
    :members:
    :inherited-members:
    :show-inheritance:

    .. autoattribute:: props
```

(python_objectmeta)=

#### ObjectMeta

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
    :special-members:
    :show-inheritance:
```

(python_datapointarray)=

#### DatapointArray

```{eval-rst}
.. autoclass:: heedy.objects.timeseries.DatapointArray
    :members:
    :show-inheritance:
```

(python_objectregistry)=

### Registering New Types

```{eval-rst}
.. automodule:: heedy.objects.registry
    :members: registerObjectType
```
