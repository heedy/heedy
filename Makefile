SRC:=$(wildcard src/core/*.go)
TMPO:=$(patsubst src/core/%.go,bin/%,$(SRC)) # Get the list of executables from the file list
OBJ:=$(TMPO:.go=)
CC:=gcc
SQLITEVERSION:=sqlite-amalgamation-3080900

.PHONY: all

all: clean bin $(OBJ) sqlite gnatsd

build: go-dependencies $(OBJ)

bin:
	mkdir bin
	cp -r config bin/config
	cp -r src/plugins/webclient/static/ bin/
	cp -r src/plugins/webclient/templates/ bin/

# Rule to go from source go file to binary
bin/%: src/core/%.go go-dependencies
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
	go get golang.org/x/tools/cmd/cover
	go get github.com/jteeuwen/go-bindata/...


gnatsd: depfolder
	go build -o bin/dep/gnatsd github.com/apcera/gnatsd

sqlite: depfolder
	unzip lib/$(SQLITEVERSION).zip -d bin/dep/
	$(CC) bin/dep/$(SQLITEVERSION)/shell.c bin/dep/$(SQLITEVERSION)/sqlite3.c -lpthread -ldl -o bin/dep/sqlite3

depfolder:
	mkdir -p bin/dep
	
# specific packages required by the project to run on a host
host-packages:
	sudo apt-get update -qq
	sudo apt-get install -qq redis-server postgresql

# run tests
test:
	./runtests.sh
