package run

/**
The plugin file specifies the interface needed to register ourselves with the
plugin registry when we're imported without side effects.
**/

import (
	"connectordb/config"
	"connectordb/plugins"
	"connectordb/security"
	"connectordb/streamdb"
	"fmt"
	"net/http"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"

	"connectordb/plugins/rest"
	"connectordb/plugins/webclient"
)

var (
	//The preferred maximum number of open files
	PreferredFileLimit = uint64(10000)
)

func init() {
	// do some sweet plugin registration!
	plugins.Register("run", usage, exec)
}

//SetFileLimit attempts to set the open file limits
func SetFileLimit() {
	var noFile syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &noFile)
	if err != nil {
		log.Warn("Could not read file limit:", err)
		return
	}
	if noFile.Cur < PreferredFileLimit {
		change := uint64(0)
		if noFile.Max < PreferredFileLimit {
			change = noFile.Max
			log.Warnf("User hard file limit (%d) is less than preferred %d", noFile.Max, PreferredFileLimit)
		} else {
			change = PreferredFileLimit
		}
		log.Warnf("Setting user file limit from %d to %d", noFile.Cur, change)
		noFile.Cur = change
		if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &noFile); err != nil {
			log.Error("Failed to set file limit: ", err)
		}
	}
}

func exec(db *streamdb.Database, args []string) error {
	SetFileLimit()

	log.Printf("Running ConnectorDB v%s on port %d", streamdb.Version, config.GetConfiguration().WebPort)
	r := mux.NewRouter()
	webclient.Setup(r, db)

	// handle the api at its versioned url
	s := r.PathPrefix("/api/v1").Subrouter()
	rest.Router(db, s)

	// all else goes to the webserver
	http.Handle("/", security.NewSecurityBuilder(r).
		//LogEverything().
		IncludeSecureHeaders().
		Hide500Errors().
		Build())

	//security.FiveHundredHandler(security.SecurityHeaderHandler(r)))

	go db.RunWriter()

	go rest.StatsRun()

	return http.ListenAndServe(fmt.Sprintf(":%d", config.GetConfiguration().WebPort), nil)
}

func usage() {
	fmt.Println(`run: Runs the full ConnectorDB system`)
}
