#!/bin/bash
echo "Waiting for servers to start..."

DBDIR="postgres_test"
PGRES_CMD="setup"
if [ -d "$DBDIR" ]; then
    PGRES_CMD="run"
fi
source ./config/runpostgres $PGRES_CMD $DBDIR
#The run script sets the global variable POSTGRES_PID to the pid of postgres
./bin/dep/gnatsd -c ./config/gnatsd.conf &
gnatsd_pid=$!
redis-server ./config/redis.conf > /dev/null 2>&1 &
redis_pid=$!
sleep 1
echo "Running tests..."
go test -cover streamdb/...
test_status=$?
kill $redis_pid
kill $gnatsd_pid
kill $POSTGRES_PID
exit $test_status
