
.PHONY: all clean build test submodules resources deps phony

all: bin/dep/gnatsd bin/connectordb resources
deps: go-dependencies submodules
build: resources bin/connectordb

#Empty rule for forcing rebuilds
phony:

bin:
	mkdir bin
	cp -r src/dbsetup/config bin/config

submodules:
	git submodule init
	git submodule update


resources: bin
	cp -r site/www/ bin/
	cp -r site/app/ bin/
	cd bin/app;bower update

# Rule to go from source go file to binary
bin/connectordb: src/connectordb.go bin phony
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
	go get github.com/tdewolff/minify
	go get golang.org/x/crypto/bcrypt
	go get github.com/dkumor/acmewrapper # Let's encrypt support

	go get github.com/stretchr/testify

	# PipeScript
	go get github.com/connectordb/pipescript


bin/dep/gnatsd: bin/dep
	go build -o bin/dep/gnatsd github.com/nats-io/gnatsd

bin/dep: bin
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
