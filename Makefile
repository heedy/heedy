GO:=go
COPY:=rsync -r --exclude=.git

.PHONY: all clean build test submodules resources deps phony

all: bin/dep/gnatsd bin/connectordb resources
deps: go-dependencies submodules
build: resources bin/connectordb

#Empty rule for forcing rebuilds
phony:

bin:
	mkdir bin
	$(COPY) src/dbsetup/config bin/

submodules:
	git submodule init
	git submodule update


resources: bin
	$(COPY) site/www bin/
	$(COPY)  site/app bin/


# Rule to go from source go file to binary
bin/connectordb: src/connectordb.go bin phony
	$(GO) build -o bin/connectordb src/connectordb.go

clean:
	rm -rf bin
	$(GO) clean


go-dependencies:
	# services
	$(GO) get github.com/nats-io/nats github.com/nats-io/gnatsd
	$(GO) get gopkg.in/redis.v3

	# databases
	$(GO) get github.com/lib/pq
	$(GO) get github.com/connectordb/duck
	$(GO) get github.com/josephlewis42/sqlx # our own so we don't depend on someone who claims the library will change in the future

	# utilities
	$(GO) get github.com/xeipuuv/gojsonschema
	$(GO) get gopkg.in/vmihailenco/msgpack.v2
	$(GO) get gopkg.in/fsnotify.v1
	$(GO) get github.com/kardianos/osext
	$(GO) get github.com/nu7hatch/gouuid
	$(GO) get github.com/gorilla/mux github.com/gorilla/context github.com/gorilla/sessions github.com/gorilla/websocket
	$(GO) get github.com/Sirupsen/logrus
	$(GO) get github.com/josephlewis42/multicache
	$(GO) get github.com/connectordb/njson
	$(GO) get github.com/codegangsta/cli
	$(GO) get github.com/tdewolff/minify
	$(GO) get golang.org/x/crypto/bcrypt
	$(GO) get github.com/dkumor/acmewrapper # Let's encrypt support

	# web services
	$(GO) get github.com/gernest/hot				# hot template reloading
	$(GO) get github.com/russross/blackfriday		# markdown processing
	$(GO) get github.com/microcosm-cc/bluemonday	# unsafe html stripper

	$(GO) get github.com/stretchr/testify

	# PipeScript
	$(GO) get github.com/connectordb/pipescript


bin/dep/gnatsd: bin/dep
	$(GO) build -o bin/dep/gnatsd github.com/nats-io/gnatsd

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
