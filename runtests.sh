#!/bin/bash
echo "Waiting for servers to start..."
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

exit $test_status
