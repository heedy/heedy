GO:=go

.PHONY: clean test phony

all: frontend streams

#Empty rule for forcing rebuilds
phony:

frontend: phony
	cd frontend; npm run build

streams: backend/main.go phony # gencode
	cd backend; $(GO) build --tags "sqlite_foreign_keys json1" -o ../assets/streams


clean:
	# $(GO) clean
	# Clear all generated assets for webapp
	rm -rf ./assets/public
	rm -f ./assets/streams