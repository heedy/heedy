GO:=go
COPY:=rsync -r --exclude=.git

.PHONY: all clean build test submodules resources deps phony

all: bin/dep/gnatsd bin/connectordb resources
deps: go-dependencies submodules app
build: resources bin/connectordb

#Empty rule for forcing rebuilds
phony:

bin:
	mkdir bin
	$(COPY) src/dbsetup/config bin/

submodules:
	git submodule update --init --recursive

app: submodules
	cd site/app;npm update

resources: bin
	$(COPY) site/www bin/
	cd site/app;npm run build


# Rule to go from source go file to binary
# http://www.atatus.com/blog/golang-auto-build-versioning/
bin/connectordb: src/main.go bin phony
	$(GO) build -o bin/connectordb -ldflags "-X commands.BuildStamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X commands.GitHash=`git rev-parse HEAD`" src/main.go

clean:
	rm -rf bin
	$(GO) clean


go-dependencies:
	# services
	$(GO) get -u github.com/nats-io/nats github.com/nats-io/gnatsd
	$(GO) get -u gopkg.in/redis.v3

	# databases
	$(GO) get -u github.com/lib/pq
	$(GO) get -u github.com/connectordb/duck
	$(GO) get -u github.com/jmoiron/sqlx

	# utilities
	$(GO) get -u github.com/xeipuuv/gojsonschema
	$(GO) get -u gopkg.in/vmihailenco/msgpack.v2
	$(GO) get -u gopkg.in/fsnotify.v1
	$(GO) get -u github.com/kardianos/osext
	$(GO) get -u github.com/nu7hatch/gouuid
	$(GO) get -u github.com/gorilla/mux github.com/gorilla/context github.com/gorilla/sessions github.com/gorilla/websocket
	$(GO) get -u github.com/Sirupsen/logrus
	$(GO) get -u github.com/josephlewis42/multicache
	$(GO) get -u github.com/connectordb/njson
	$(GO) get -u https://github.com/spf13/cobra
	$(GO) get -u github.com/tdewolff/minify
	$(GO) get -u golang.org/x/crypto/bcrypt
	$(GO) get -u github.com/dkumor/acmewrapper # Let's encrypt support

	# web services
	$(GO) get -u github.com/gernest/hot				# hot template reloading
	$(GO) get -u github.com/russross/blackfriday		# markdown processing
	$(GO) get -u github.com/microcosm-cc/bluemonday	# unsafe html stripper

	$(GO) get -u github.com/stretchr/testify

	# PipeScript
	$(GO) get -u github.com/connectordb/pipescript


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
