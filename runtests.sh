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

echo "My PID"

echo $$

check_pids () { 
	echo "==================================================="
	echo "Checking For Runaway Processes"
	echo "==================================================="

	echo "Looking for postgres proc..."
	ps aux | grep postgres | grep -v 'grep'

	echo "Looking for redis proc..."
	ps aux | grep redis-server | grep -v 'grep'

	echo "Looking for gnatsd proc..."
	ps aux | grep gnatsd | grep -v 'grep'
	
	echo "Looking for connectordb proc..."
	ps aux | grep connectordb | grep -v 'grep'
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
	./bin/connectordb create $DBDIR -user=test:test
}

force_stop

create

check_pids

stop
check_pids
force_stop

start
check_pids

echo "==================================================="
echo "Running coverage tests"
echo "==================================================="
go test -cover streamdb/...
test_status=$?

stop
check_pids
force_stop

if [ $test_status -eq 0 ]; then
    rm -rf $DBDIR
    create
    check_pids
    stop
    check_pids
    force_stop
    
    start
    check_pids
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
	./bin/connectordb stop $DBDIR
fi


#rm -rf $DBDIR
exit $test_status
