#!/bin/bash
./bin/gnatsd -c ./config/gnatsd.conf &
gnatsd_pid=$!
redis-server ./config/redis.conf > /dev/null 2>&1 &
redis_pid=$!
echo "Waiting for servers to start..."
sleep 20
echo "Running tests..."
go test streamdb/...
test_status=$?
kill $redis_pid
kill $gnatsd_pid
exit $test_status
