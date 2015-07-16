ConnectorDB Python Interface
=====

This is a mini-crash-course in the ConnectorDB python interface

To install:
```python
sudo python setup.py install
```

If on ubuntu, you might want to install `python-apsw` before running the above command.
You can look at the dependencies in setup.py.

Logging
----------------

The simplest, and most common task is logging data. Suppose we have a little weather station that gives us the temperature every minute.

```python

def getTemperature():
	#Your code here

from connectordb.logger import ConnectorLogger

#Open the logger using a cache file name (where datapoints are cached before syncing)
l = ConnectorLogger("cache.db")

#Set the device you are using
l.setLogin("myname/weatherstation","apikeysagr43ETE343were")

#If given a schema (as we have done here), addStream will create the stream if it doesn't exist
l.addStream("temperature",{"type":"number"})

#Sync with ConnectorDB once an hour (in seconds)
l.syncperiod = 60*60

#Start running syncer in background
l.start()

#While the syncer is running in the background, we are free to add data
#to the cache however we want - it will be saved first to the cache file
#so that you don't lose any data, and will be synced to the database once an hour
while True:
	time.sleep(60)
	l.insert("temperature",getTemperature())
```

ConnectorDB Basics
--------------------

The logger is a convenient wrapper for gathering data. When wanting to operate on the database directly, you will want to use the `ConnectorDB` object:

```python
from connectordb import ConnectorDB

db = ConnectorDB("myname/mydevice","apikeysadfdsf98439g")

mystream = db["mystream"]

if not mystream.exists():
	mystream.create({"type": "string"})

mystream.insert("Hello World!")

#mystream has 1 datapoint
print "mystream has",len(mystream),"datapoint"

#Prints Hello World! - each datapoint is a dict where "t" is the timestamp, and "d" is the data
print mystream[0]["d"]
```

if you need super fast inserts, you can enable websocket inserts:

```python
db.wsinsert = True
```

which will make `db.insert` use websocket.
Beware, though: inserting thru websocket does not return errors!

Subscriptions
-------------------

You can subscribe to streams, so that you get data the moment it is written to the database.

```python
from connectordb import ConnectorDB

db = ConnectorDB("myname/mydevice","apikeysadfdsf98439g")

mystream = db["mystream"]

if not mystream.exists():
	mystream.create({"type": "string"})

def callbackFunction(streampath,data):
	print streampath,data

mystream.subscribe(callbackFunction)

#After inserting, callbackFunction will run!
mystream.insert("Hello World!")

```

If you are implementing a downlink stream (a stream that accepts input from other devices, such as a light switch), subscribe to the downlink:

```python
from connectordb import ConnectorDB
import time

db = ConnectorDB("myname/mydevice","apikeysadfdsf98439g")

mystream = db["mylight"]

if not mystream.exists():
	mystream.create({"type": "boolean"})
	mystream.downlink = True

def callbackFunction(streampath,data):
	ison = data[-1]["d"]
	if ison:
		turn_on_light()
	else:
		turn_off_light()
	#This acknowledges the datapoint by writing the action that was taken to the real stream
	#As a shortcut, you can also use return True to acknowledge the unmodified data
	#meaning that return True would return the original data as given by the data variable.
	#Not returning anything or returning False does not acknowledge the downlink.
	#WARNING: Make sure only ONE callback acknowledges the downlink to avoid double-inserts
	#DANGER: Make sure you only acknowledge downlinks to streams that belong to the currently
	#authenticated device, since inserting as a different device will redirect to
	#the downlink stream, creating an infinite insert loop.
	return [{"t": time.time(),"d": ison}]

mystream.subscribe(callbackFunction,downlink=True)

#Sleep forever waiting for inputs. You can also just do an infinite loop with sleep.
#in fact, that's what sleepforever does
db.sleepforever()
```
