#!/bin/bash

./bin/connectordb --version

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
	killall postgres redis-server gnatsd connectordb

    sleep 1

	killall -9 postgres redis-server gnatsd connectordb
}

start () {
	echo "==================================================="
	echo "Doing Backend Start"
	echo "==================================================="
	./bin/connectordb start $DBDIR --backend
}

create () {
	echo "==================================================="
	echo "Doing Create"
	echo "==================================================="
	rm -rf $DBDIR
	./bin/connectordb create $DBDIR --test
}

force_stop

create
check_pids

start

echo "==================================================="
echo "Running coverage tests"
echo "==================================================="
#go test --timeout 15s -p=1 -v -cover connectordb/...
go test --timeout 500s -p=1 -cover connectordb/...
test_status=$?
if [ "$test_status" -ne 0 ]; then
    stop
    exit $test_status
fi
go test --timeout 15s -p=1 -cover util/...
test_status=$?
if [ "$test_status" -ne 0 ]; then
    stop
    exit $test_status
fi
go test --timeout 15s -p=1 -cover server/...
test_status=$?
if [ "$test_status" -ne 0 ]; then
    stop
    exit $test_status
fi

#go test --timeout 15s -p=1 -bench . connectordb/...
#test_status=$?
stop
check_pids

if [[ $1 == "coveronly" ]]; then
    exit 0
fi


create

#Now test the python stuff, while rebuilding the db to make sure that
#the go tests didn't invalidate the db
echo "==================================================="
echo "Starting Server"
echo "==================================================="
./bin/connectordb -l=DEBUG start $DBDIR

echo "==================================================="
echo "Starting API Tests"
echo "==================================================="
nosetests --with-coverage --cover-package=connectordb -s --nologcapture connectordb_python/connectordb_test.py connectordb_python/query_test.py
test_status=$?

stop

check_pids

#delete dir if tests succeeded, and then redo the api tests on sqlite
if [ "$test_status" -eq 0 ]; then
    rm -rf $DBDIR

    ./bin/connectordb -l=DEBUG create $DBDIR --test --sqlbackend=sqlite3
    ./bin/connectordb -l=DEBUG start $DBDIR
    nosetests --with-coverage --cover-package=connectordb -s --nologcapture connectordb_python/connectordb_test.py connectordb_python/query_test.py
    test_status=$?
    stop
    check_pids

    if [ "$test_status" -eq 0 ]; then
        rm -rf $DBDIR
    fi

fi

exit $test_status
