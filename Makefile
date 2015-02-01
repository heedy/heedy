.PHONY: clean all dependencies test gnatsd

#gets the list of files that we're to compile
SRC=$(wildcard tools/*.go)

#Get the list of executables from the file list
TMPO=$(patsubst tools/%.go,bin/%,$(SRC))
OBJ=$(TMPO:.go=)

#Rule to go from source go file to binary
bin/%: tools/%.go bin
	go build -o $@ $<

all: $(OBJ)
bin:
	mkdir bin
clean:
	rm -rf bin

############################################################################################################
#Dependencies of the project
############################################################################################################

dependencies: bin
	go get github.com/apcera/nats
	go get github.com/apcera/gnatsd
	go get github.com/garyburd/redigo/redis
	go get github.com/mattn/go-sqlite3
	go get github.com/nu7hatch/gouuid
	go get github.com/gorilla/mux
	go get github.com/gorilla/context
	go build -o bin/gnatsd github.com/apcera/gnatsd

############################################################################################################
#Running Tests
############################################################################################################

test:
	./runtests.sh
