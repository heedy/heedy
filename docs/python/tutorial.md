# Tutorial

When using the Python library, you will either be calling code using an App access token, in which case you can log in with:

```python
import heedy
h = heedy.App("your_access_token","https://heedy.mydomain.com")
```

or, if you are writing a heedy plugin, you can use:

```python
import heedy
p = heedy.Plugin()
```

Unless otherwise specified, the rest of this tutorial assumes that the `h` variable refers to an App object.

## Basics

To start off, let's check which objects belong to this app

```python
>>> h.objects()
[]
```

Apparently, the app does not have any objects yet. Let's create a timeseries object,
into which we will be able to add data:

```python
>>> ts = h.objects.create("My Timeseries",tags="mytimeseries")
>>> ts
Timeseries{'access': '*',
 'app': 'cead7c85-f992-4a56-898c-ec44c8322d5c',
 'created_date': '2020-03-09',
 'description': '',
 'id': 'a2f0ae8b-73e7-4ffa-b078-a1ca3f260437',
 'modified_date': None,
 'meta': {'actor': False, 'schema': {}, 'subtype': ''},
 'name': 'My Timeseries',
 'owner': 'test',
 'owner_scope': '*',
 'tags': 'mytimeseries',
 'type': 'timeseries'}
```

```eval_rst
.. note::
    If you get an access denied error, make sure your app has the `self.objects` permission scope.
```

Notice that the timeseries object has an access of `*`. This means that our app has full access to the object. The app owner (the `test` user) is also given full access to the timeseries (`owner_scope` is `*`). Let's disallow the user from adding data, since our app will manage this timeseries:

```python
>>> ts.owner_scope = "read"
```

The same effect can be achieved with `ts.update(owner_scope="read")`, which is awaitable when using an async session.

## Querying Objects

In the previous section, we gave the object a name, as well as tags. Since object IDs can be difficult to remember, each object can have a set of tags (space separated), and can be queried using them:

```python
>>> h.objects(tags="mytimeseries")
```

Alternately, since we are creating objects for our own app, we could have set an object key. While multiple objects can share tags, an object key is unique for the app's objects:

```python
>>> ts.key = "mytimeseries"
>>> h.objects(key="mytimeseries")
```

## Timeseries

Our object was of type `timeseries`. This means that it has special functionality. A heedy timeseries object can be accessed like an array in Python:

```python
>>> len(ts)
0
>>> ts.append(5)
{'result': 'ok'}
>>> len(ts)
1
>>> ts[:]
[{'t': 1583735182.6548438, 'd': 5}]
```

When data is appended to the timeseries, it is given the current unix timestamp. This means that the timeseries can be accessed using time-based queries:

```python
>>> ts(t1="now-1h") # Get the most recent 1 hour of data
>>> ts(t1=1583735182,t2=1583735183) # Get data between unix timestamps
```

Currently, the timeseries can accept any type of data. We want it to only accept numbers.
This can be achieved by giving it a [JSON Schema](https://json-schema.org/).
Since the schema property is specific to objects of the timeseries type, it is located in the object metadata. You can set elements of the metadata by creating an object which will overwrite the desired properties:

```python
>>> ts.meta = {"schema": {"type":"number"}}
>>> ts
Timeseries{'access': '*',
 ...
 'meta': {'actor': False, 'schema': {"type":"number"}, 'subtype': ''},
 ...}
>>> ts.append("I am a string")
Traceback (most recent call last):
  File "<stdin>", line 1, in <module>
  ...
  File "/usr/lib/python3.8/site-packages/heedy/base.py", line 127, in handleResponse
    raise HeedyError(msg)
heedy.base.HeedyError: bad_query: The data failed schema validation
>>> ts.append(3.14)
{'result': 'ok'}
```

## Hybrid Asyncio

Python 3 has support for [asynchronous operations with asyncio](https://docs.python.org/3/library/asyncio.html), which are incompatible with standard blocking calls.
While most analysis code will use blocking code, web servers (and therefore heedy plugins) will often want to use an asynchronous version of a given library.

Rather than maintaining two separate libraries, heedy's Python support was built as a _hybrid_, which can be used in _both_ situations. You simply need to notify which version you will be using when loading your app:

```python
# Blocking (sync) mode (default for apps)
h = heedy.App("your_access_token",session="sync")
print(h.objects()) # List the objects with a blocking call

# Non-blocking (async) mode
h = heedy.App("your_access_token",session="async")
print(await h.objects()) # Asynchronous object listing
```
