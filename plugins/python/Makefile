GO:=go

.PHONY: clean test phony

all: 

#Empty rule for forcing rebuilds
phony:


server: backend/main.go phony # gencode
	cd backend; $(GO) build --tags "sqlite_foreign_keys json1 sqlite_preupdate_hook" -o ../assets/server

standalone: server

clean:
	# $(GO) clean
	rm -f ./assets/server