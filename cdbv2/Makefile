GO:=go

.PHONY: clean test phony

all: connectordb

#Empty rule for forcing rebuilds
phony:

connectordb: src/main.go phony
	cd src; $(GO) build -o ../connectordb

clean:
	$(GO) clean