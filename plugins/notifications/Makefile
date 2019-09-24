GO:=go

.PHONY: clean test phony

all: 

#Empty rule for forcing rebuilds
phony:

frontend: phony
	cd frontend; npm run build

server: backend/main.go phony # gencode
	cd backend; $(GO) build --tags "sqlite_foreign_keys json1 sqlite_preupdate_hook" -o ../assets/server

standalone: server frontend

builtin:
	rm -rf ./assets/public
	ln -s ../../../assets/public assets/public
	
debug:
	cd frontend; npm run mkdebug

clean:
	# $(GO) clean
	# Clear all generated assets for webapp
	rm -rf ./assets/public
	rm -f ./assets/server