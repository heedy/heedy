GO:=go

.PHONY: clean test phony

all: setup app connectordb

#Empty rule for forcing rebuilds
phony:

setup: phony
	cd setup; npm run build

app: phony
	cd app; npm run build

connectordb: src/main.go phony
	cd src; GO111MODULE=on packr2
	cd src; $(GO) build -o ../connectordb
	cd src; GO111MODULE=on packr2 clean

debug:
	cd src; $(GO) build -o ../connectordb

clean:
	$(GO) clean
	rm -rf ./assets/app
	rm -rf ./assets/setup
	cd src; GO111MODULE=on packr2 clean