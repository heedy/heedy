#!/bin/bash
echo "Waiting for servers to start..."

DBDIR="database_test"

if [ "$EUID" -eq 0 ]
  then echo "Do not run this script as root!"
  exit 1
fi


if [ -d "$DBDIR" ]; then
    rm -rf $DBDIR
fi

killall postgres
killall gnatsd
killall redis-server

echo "Setting up environment..."
export PATH=bin/dep:$PATH

./bin/connectordb create $DBDIR
killall postgres
killall gnatsd
killall redis-server
./bin/connectordb start $DBDIR servers &

sleep 4

echo "Running tests..."
go test -cover streamdb/...
test_status=$?

./bin/connectordb stop $DBDIR
rm -rf $DBDIR

killall connectordb
killall postgres
killall gnatsd
killall redis-server

if [ $test_status -eq 0 ]; then
	#Now test the python stuff, while rebuilding the db to make sure that
	#the go tests didn't invalidate the db
	./bin/connectordb create $DBDIR --user test:test

	killall postgres
	killall gnatsd
	killall redis-server

	./bin/connectordb start $DBDIR &


    echo "Starting connectordb api tests..."
    nosetests src/clients/python/connectordb_test.py
    test_status=$?
    
	./bin/connectordb stop $DBDIR

	killall connectordb
	killall postgres
	killall gnatsd
	killall redis-server
fi


if [ $test_status -eq 0 ]; then
	rm -rf $DBDIR
fi
exit $test_status
