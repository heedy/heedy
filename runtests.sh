#!/bin/bash
echo "Waiting for servers to start..."

#Make sure MongoDB is running
mongo --eval "db.stats()"
mongoresult=$?
mongo_pid=0
if [ $mongoresult -ne 0 ]; then
    mkdir mongodb_database
    mongod --config ./config/mongo.conf &
    mongo_pid=$!
    sleep 5
fi

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

if [ $mongo_pid -ne 0 ]; then
    kill $mongo_pid
fi


exit $test_status
