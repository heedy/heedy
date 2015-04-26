SRC:=$(wildcard src/core/*.go)
TMPO:=$(patsubst src/core/%.go,bin/%,$(SRC)) # Get the list of executables from the file list
OBJ:=$(TMPO:.go=)
CC:=gcc

.PHONY: all clean build

all: go-dependencies $(OBJ) bin/dep/gnatsd

build: $(OBJ)

bin:
	mkdir bin
	cp -r config bin/config
	cp -r src/plugins/web_client/static/ bin/
	cp -r src/plugins/web_client/templates/ bin/

# Rule to go from source go file to binary
bin/%: src/core/%.go bin
	go build -o $@ $<

clean:
	rm -rf bin
	go clean


go-dependencies:
	# services
	go get github.com/apcera/nats github.com/apcera/gnatsd
	go get github.com/garyburd/redigo/redis

	# databases
	go get github.com/lib/pq
	go get github.com/mattn/go-sqlite3
	go get github.com/josephlewis42/sqlx # our own so we don't depend on someone who claims the library will change in the future

	# utilities
	go get github.com/xeipuuv/gojsonschema
	go get github.com/vharitonsky/iniflags
	go get github.com/kardianos/osext
	go get github.com/nu7hatch/gouuid
	go get github.com/gorilla/mux github.com/gorilla/context github.com/gorilla/sessions

	# go tools
	go get github.com/jteeuwen/go-bindata/...


bin/dep/gnatsd: depfolder
	go build -o bin/dep/gnatsd github.com/apcera/gnatsd

depfolder:
	mkdir -p bin/dep

# specific packages required by the project to run on a host
host-packages:
	sudo apt-get update -qq
	sudo apt-get install -qq redis-server postgresql

# run tests
test:
	./runtests.sh
