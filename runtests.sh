#!/bin/bash
echo "Waiting for servers to start..."

DBDIR="postgres_test"

if [ "$EUID" -eq 0 ]
  then echo "Do not run this script as root!"
  exit 1
fi


killall postgres
if [ -d "$DBDIR" ]; then
    rm -rf $DBDIR
fi

# we setup every time for a clean environment
source ./config/runpostgres setup $DBDIR

#The run script sets the global variable POSTGRES_PID to the pid of postgres
./bin/dep/gnatsd -c ./config/gnatsd.conf &
gnatsd_pid=$!
redis-server ./config/redis.conf > /dev/null 2>&1 &
redis_pid=$!
sleep 1
echo "Running tests..."
go test -cover streamdb/...
test_status=$?

if [ $test_status -eq 0 ]; then
    ./bin/restserver &
    rest_pid=$!
    sleep 1
    nosetests src/clients/python/connectordb_test.py
    test_status=$?
    kill $rest_pid
fi

kill $redis_pid
kill $gnatsd_pid
kill $POSTGRES_PID
exit $test_status
