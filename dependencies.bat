:: These are basically explicit copies of the go get stuff in the makefile
:: The linux version is source of truth.
go get -u gopkg.in/redis.v3
go get -u github.com/nats-io/nats
go get -u github.com/lib/pq
go get -u github.com/connectordb/duck
go get -u github.com/jmoiron/sqlx
go get -u github.com/xeipuuv/gojsonschema
go get -u gopkg.in/vmihailenco/msgpack.v2
go get -u gopkg.in/fsnotify.v1
go get -u github.com/kardianos/osext
go get -u github.com/nu7hatch/gouuid
go get -u github.com/gorilla/mux github.com/gorilla/context github.com/gorilla/sessions github.com/gorilla/websocket
go get -u github.com/Sirupsen/logrus
go get -u github.com/josephlewis42/multicache
go get -u github.com/connectordb/njson
go get -u github.com/spf13/cobra
go get -u github.com/tdewolff/minify
go get -u golang.org/x/crypto/bcrypt
go get -u github.com/dkumor/acmewrapper
go get -u github.com/gernest/hot
go get -u github.com/russross/blackfriday
go get -u github.com/microcosm-cc/bluemonday
go get -u github.com/stretchr/testify
go get -u github.com/connectordb/pipescript


if not exist "bin" mkdir "bin"
robocopy /s "./src/dbsetup/config" "./bin/config"

exit /B
