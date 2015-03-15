#!/bin/bash
echo "Waiting for servers to start..."

DBDIR="../postgres_production"

if [ "$EUID" -eq 0 ]
  then echo "Do not run this script as root!"
  exit 1
fi


# we setup every time for a clean environment
./config/runpostgres setup $DBDIR
./config/runpostgres run $DBDIR

#The run script sets the global variable POSTGRES_PID to the pid of postgres
./bin/dep/gnatsd -c ./config/gnatsd.conf &
gnatsd_pid=$!
redis-server ./config/redis.conf > /dev/null 2>&1 &
redis_pid=$!
sleep 1

#./bin/webservice --sql "sslmode=disable dbname=connectordb port=52592"
./bin/webservice --sql "database.sqlite3"
