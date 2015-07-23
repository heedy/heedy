SRC:=$(wildcard src/core/*.go)
TMPO:=$(patsubst src/core/%.go,bin/%,$(SRC)) # Get the list of executables from the file list
OBJ:=$(TMPO:.go=)
CC:=gcc

.PHONY: all clean build test

all: resources go-dependencies $(OBJ) bin/dep/gnatsd

build: resources $(OBJ)

bin:
	mkdir bin
	cp -r src/connectordb/services/config bin/config

resources: bin
	cp -r src/connectordb/plugins/webclient/static/ bin/
	cp -r src/connectordb/plugins/webclient/templates/ bin/

# Rule to go from source go file to binary
bin/%: src/core/%.go bin go-dependencies
	go build -o $@ $<

clean:
	rm -rf bin
	go clean


go-dependencies:
	# services
	go get github.com/nats-io/nats github.com/nats-io/gnatsd
	go get gopkg.in/redis.v3

	# databases
	go get github.com/lib/pq
	go get github.com/josephlewis42/sqlx # our own so we don't depend on someone who claims the library will change in the future

	# utilities
	go get github.com/xeipuuv/gojsonschema
	go get gopkg.in/vmihailenco/msgpack.v2
	go get github.com/vharitonsky/iniflags
	go get github.com/kardianos/osext
	go get github.com/nu7hatch/gouuid
	go get github.com/gorilla/mux github.com/gorilla/context github.com/gorilla/sessions github.com/gorilla/websocket github.com/gorilla/feeds
	go get github.com/Sirupsen/logrus
	go get github.com/josephlewis42/multicache
	go get github.com/connectordb/njson

	go get github.com/stretchr/testify


bin/dep/gnatsd: bin/dep go-dependencies
	go build -o bin/dep/gnatsd github.com/nats-io/gnatsd

bin/dep:
	mkdir -p bin/dep

# specific packages required by the project to run on a host
host-packages:
	sudo apt-get update -qq
	sudo apt-get install -qq redis-server postgresql

# run tests
test:
	./runtests.sh
