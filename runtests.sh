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

echo "Setting up environment..."

./bin/connectordb2 create $DBDIR
./bin/connectordb2 start $DBDIR &

echo "Running tests..."
go test -cover streamdb/...
test_status=$?



if [ $test_status -eq 0 ]; then
    echo "Starting connectordb api tests..."
    nosetests src/clients/python/connectordb_test.py
    test_status=$?
fi
exit $test_status
