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

echo "My PID" $$
test_status=0
pg_pre=`ps aux | grep postgres | grep -v grep | awk '{print $2}'`
redis_pre=`ps aux | grep redis-server | grep -v grep | awk '{print $2}'`
gnatsd_pre=`ps aux | grep gnatsd | grep -v grep | awk '{print $2}'`

check_pids () {
	echo "==================================================="
	echo "Checking For Runaway Processes"
	echo "==================================================="


	echo "Looking for postgres proc..."
	postgresproc=`ps aux | grep postgres | grep -v grep | awk '{print $2}'`
	echo $postgresproc

	echo "Looking for redis proc..."
	redisproc=`ps aux | grep redis-server | grep -v grep | awk '{print $2}'`
	echo $redisproc

	echo "Looking for gnatsd proc..."
	gnatsdproc=`ps aux | grep gnatsd | grep -v grep | awk '{print $2}'`
	echo $gnatsdproc

	echo "Looking for connectordb proc..."
	ps aux | grep connectordb | grep -v 'grep'

	if [ "$postgresproc" != "$pg_pre" ]; then
		echo "Postgres process started from us was still running"
		exit 1
	fi

	if [ "$redisproc" != "$redis_pre" ]; then
		echo "Redis process started from us was still running"
		exit 1
	fi

	if [ "$gnatsdproc" != "$gnatsd_pre" ]; then
		echo "Gnatsd process started from us was still running"
		exit 1
	fi

    if [ "$test_status" != 0 ]; then
        echo "FAILED TEST"
        exit 1
    fi
}

stop () {
	echo "==================================================="
	echo "Doing Stop"
	echo "==================================================="
	./bin/connectordb stop $DBDIR
}

force_stop () {
	killall postgres
	killall redis-server
	killall gnatsd
	killall connectordb
}

start () {
	echo "==================================================="
	echo "Doing Start"
	echo "==================================================="
	./bin/connectordb start $DBDIR
}

create () {
	echo "==================================================="
	echo "Doing Create"
	echo "==================================================="
	rm -rf $DBDIR
	./bin/connectordb create $DBDIR -user=test:test
}

force_stop

create
check_pids

start

echo "==================================================="
echo "Running coverage tests"
echo "==================================================="
go test -cover connectordb/...
test_status=$?
stop
check_pids

create
start
#Now test the python stuff, while rebuilding the db to make sure that
#the go tests didn't invalidate the db
echo "==================================================="
echo "Starting Rest"
echo "==================================================="
./bin/restserver --sql=postgres://127.0.0.1:52592/connectordb?sslmode=disable &
rest_server=$!

echo "==================================================="
echo "Starting API Tests"
echo "==================================================="
nosetests src/clients/python/connectordb_test.py
test_status=$?

kill $rest_server
stop

check_pids

#delete dir if tests succeeded
if [ "$test_status" -eq 0 ]; then
    rm -rf $DBDIR
fi

exit $test_status
