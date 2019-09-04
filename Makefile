GO:=go

VERSION:=$(shell cat VERSION)-git.$(shell git rev-list --count HEAD)

.PHONY: clean test phony

all: frontend heedy

#Empty rule for forcing rebuilds
phony:

frontend: phony
	cd frontend; npm run build

heedy: backend/main.go phony # gencode
	statik -src=./assets -dest=./backend -p assets -f
	cd backend; $(GO) build --tags "sqlite_foreign_keys json1" -o ../heedy -ldflags "-X \"github.com/heedy/heedy/backend/buildinfo.BuildTimestamp=`date -u '+%Y-%m-%d %H:%M:%S'`\" -X github.com/heedy/heedy/backend/buildinfo.GitHash=`git rev-parse HEAD` -X github.com/heedy/heedy/backend/buildinfo.Version=$(VERSION)"
	rm ./backend/assets/statik.go

debug:
	cd backend; $(GO) build --tags "sqlite_foreign_keys json1" -o ../heedy -ldflags "-X \"github.com/heedy/heedy/backend/buildinfo.BuildTimestamp=`date -u '+%Y-%m-%d %H:%M:%S'`\" -X github.com/heedy/heedy/backend/buildinfo.GitHash=`git rev-parse HEAD` -X github.com/heedy/heedy/backend/buildinfo.Version=`cat ../VERSION`-debug.`git rev-list --count HEAD`"

clean:
	# $(GO) clean
	# Clear all generated assets for webapp
	rm -rf ./assets/public
	rm -f heedy
	# Clean docs
	# cd docs/en; make clean

	# Clear any assets packed by statik
	rm -f ./backend/assets/statik.go
