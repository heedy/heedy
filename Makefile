
.PHONY: all clean build test

all: resources go-dependencies bin/dep/gnatsd bin/connectordb

build: app resources bin/connectordb

bin:
	mkdir bin
	cp -r src/dbsetup/config bin/config

app:
	git submodule init
	git submodule update
	cd site/app;bower update


resources: bin app
	cp -r site/www/ bin/
	cp -r site/app/ bin/

# Rule to go from source go file to binary
bin/connectordb: src/connectordb.go bin go-dependencies
	go build -o bin/connectordb src/connectordb.go

clean:
	rm -rf bin
	go clean


go-dependencies:
	# services
	go get github.com/nats-io/nats github.com/nats-io/gnatsd
	go get gopkg.in/redis.v3

	# databases
	go get github.com/lib/pq
	go get github.com/connectordb/duck
	go get github.com/josephlewis42/sqlx # our own so we don't depend on someone who claims the library will change in the future

	# utilities
	go get github.com/xeipuuv/gojsonschema
	go get gopkg.in/vmihailenco/msgpack.v2
	go get gopkg.in/fsnotify.v1
	go get github.com/vharitonsky/iniflags
	go get github.com/kardianos/osext
	go get github.com/nu7hatch/gouuid
	go get github.com/gorilla/mux github.com/gorilla/context github.com/gorilla/sessions github.com/gorilla/websocket
	go get github.com/Sirupsen/logrus
	go get github.com/josephlewis42/multicache
	go get github.com/connectordb/njson
	go get github.com/codegangsta/cli

	go get github.com/stretchr/testify


bin/dep/gnatsd: bin/dep go-dependencies
	go build -o bin/dep/gnatsd github.com/nats-io/gnatsd

bin/dep:
	mkdir -p bin/dep

# specific packages required by the project to run on a host
host-packages:
	sudo apt-get update -qq
	sudo apt-get install -qq redis-server postgresql

connectordb_python:
	git clone https://github.com/connectordb/connectordb_python

# run tests
test: connectordb_python
	./runtests.sh
