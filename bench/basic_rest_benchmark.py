#!/usr/bin/env python2

''' This script provides an end to end benchmark for ConnectorDB.

Due to the GIL it may be prudent to write this in Java once we have a client
for it.

Copyright 2015 Joseph Lewis <joseph@josephlewis.net>
All Rights Reserved

'''


import connectordb
import threading
import time

NUM_USERS = 10
DEVICES_PER_USER = 5
STREAMS_PER_DEVICE = 10
POINTS_PER_STREAM = 100

ADMIN_USERNAME = "admin"
ADMIN_PASSWORD = "admin"
HOST = "http://127.0.0.1:8000"

# use a suffix for names so we can run multiple times on a single DB
SUFFIX = str(int(time.time()))


# We create all users before we do anything else.
c = connectordb.ConnectorDB(ADMIN_USERNAME, ADMIN_PASSWORD, HOST)

users = ["test_{}_{}".format(i, SUFFIX) for i in xrange(NUM_USERS)]

# Create all users
for uname in users:
	u = c(uname)
	u.create("{}@localhost".format(uname), uname)



def benchmark(uname):
	c = connectordb.ConnectorDB(uname, uname, HOST)
	
	devices = ["{}/{}".format(uname, i) for i in xrange(DEVICES_PER_USER)]
	stream_suffixes = ["/" + str(i) for i in xrange(STREAMS_PER_DEVICE)]
	streams = []
	for dev in devices:
		for suffix in stream_suffixes:
			streams.append(dev + suffix)
	
	# Create all needed objects
	for dev in devices:
		c(dev).create()
	
	streams = [c(stream) for stream in streams]
	for stream in streams:
		stream.create({"type":"number"})
	
	for i in xrange(POINTS_PER_STREAM * STREAMS_PER_DEVICE * DEVICES_PER_USER):
		streams[i % len(streams)].insert(i)
		
	
# Start all the users going
threads = []
for uname in users:
	benchmark(uname)
	threads.append(threading.Thread(target=benchmark, args=(uname,)))

for thread in threads:
	thread.start()

for thread in threads:
	thread.join()
