GO:=go

.PHONY: clean test phony

all: setup frontend heedy

#Empty rule for forcing rebuilds
phony:

setup: phony
	cd setup; npm run build

frontend: phony
	cd frontend; npm run build

heedy: backend/main.go phony # gencode
	statik -src=./assets -dest=./backend -p assets -f
	cd backend; $(GO) build --tags "sqlite_foreign_keys" -o ../heedy
	rm ./backend/assets/statik.go

debug: #gencode
	cd backend; $(GO) build --tags "sqlite_foreign_keys" -o ../heedy

clean:
	# $(GO) clean
	# Clear all generated assets for webapp
	rm -rf ./assets/public
	rm -rf ./assets/setup
	rm -rf heedy
	# Clean docs
	cd docs/en; make clean

	# Clear any assets packed by statik
	rm -f ./backend/assets/statik.go