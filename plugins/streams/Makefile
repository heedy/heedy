GO:=go

.PHONY: clean test phony

all: 

#Empty rule for forcing rebuilds
phony:

frontend: phony
	cd frontend; npm run build

streams: backend/main.go phony # gencode
	cd backend; $(GO) build --tags "sqlite_foreign_keys json1" -o ../assets/streams

standalone: streams

builtin:
	rm -rf ./assets/public
	ln -s ../../../assets/public assets/public
	
debug:
	cd frontend; npm run debug

clean:
	# $(GO) clean
	# Clear all generated assets for webapp
	rm -rf ./assets/public
	rm -f ./assets/streams