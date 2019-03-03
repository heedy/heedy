GO:=go

.PHONY: clean test phony

all: setup frontend connectordb

#Empty rule for forcing rebuilds
phony:

setup: phony
	cd setup; npm run build

frontend: phony
	rm -rf src/api/proto;
	cd frontend; npm run build

docs: phony
	protoc -I ./src/api/ -I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway api.proto --swagger_out=logtostderr=true:docs
	cd docs; make html

#gencode: phony
#	mkdir -p src/api/pb
#	protoc -I ./src/api/ -I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway api.proto --go_out=plugins=grpc:src/api/pb --grpc-gateway_out=logtostderr=true:src/api/pb

connectordb: src/main.go phony # gencode
	statik -src=./assets -dest=./src -p assets -f
	cd src; $(GO) build --tags "sqlite_foreign_keys" -o ../connectordb
	rm ./src/assets/statik.go

debug: #gencode
	cd src; $(GO) build --tags "sqlite_foreign_keys" -o ../connectordb

clean:
	# $(GO) clean
	# Clear all generated assets for webapp
	rm -rf ./assets/public
	rm -rf ./assets/setup
	# Remove the generated APIs
	rm -rf src/api/proto
	rm -rf docs/api.swagger.json
	rm -rf connectordb
	# Clean docs
	cd docs/en; make clean

	# Clear any assets packed by statik
	rm -f ./src/assets/statik.go