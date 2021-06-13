package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/database/dbutil"
	"github.com/heedy/heedy/backend/plugins/run"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Plugin contains methods that can be used when generating golang heedy plugins
// it is the main interface with the main server
type Plugin struct {
	Meta *run.Info
	ADB  *database.AdminDB
}

// Init is to be run right at the start of the plugin, and it can only be run once.
// It parses the information incoming from heedy, and prepares the relevant methods
func Init() (*Plugin, error) {
	var ex run.Info
	reader := bufio.NewReader(os.Stdin)
	b, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &ex)
	if err != nil {
		return nil, err
	}

	// TODO: set log level from the config
	log.SetLevel(log.DebugLevel)

	return &Plugin{
		Meta: &ex,
	}, nil
}

func (p *Plugin) String() string {
	b, _ := json.Marshal(p.Meta)
	return string(b)
}

// Logger returns a logger built to be compatible with heedy
func (p *Plugin) Logger() *log.Entry {

	return log.WithField("run", p.Meta.Plugin+":"+p.Meta.Name)
}

// AdminDB returns the heedy adminsitrative database. Through the AdminDB, the plugin can
// make direct sql queries to the database. Be aware that the adminDB does not go through
// the heedy server at all - it operates directly upon the sql database. For this reason,
// it is recommended to use PluginDB for queries that can be handled by the heedy server,
// since the plugin version might not be exactly aligned with the server version, and
// it might cause compatibility issues. AdminDB is best used for raw sql queries to the database.
func (p *Plugin) AdminDB() (*database.AdminDB, error) {
	if p.ADB != nil {
		return p.ADB, nil
	}
	db, err := database.Open(&assets.Assets{
		FolderPath: p.Meta.HeedyDir,
		Config:     p.Meta.Config,
	})
	p.ADB = db
	return db, err
}

// Returns the PluginDB acting as the given entity
func (p *Plugin) As(entity string) *PluginDB {

	method, host, err := run.GetEndpoint(p.Meta.DataDir, p.Meta.Config.GetAPI())
	if err != nil {
		logrus.Panicf("Got an error parsing config API %v at %s", err, dbutil.MiniStack(0))
	}

	c := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	if method == "unix" {
		c.Transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", host)
			},
		}
	}

	return &PluginDB{
		P:      p,
		host:   host,
		client: c,
		Entity: entity,
	}
}

func (p *Plugin) Close() {
	if p.ADB != nil {
		p.ADB.Close()
	}
}

// InitSQL initializes the plugin's sql portion
func (p *Plugin) InitSQL(name string, version int, updater func(*database.AdminDB, *run.Info, run.BuiltinHelper, int) error) error {
	adb, err := p.AdminDB()
	if err != nil {
		return err
	}

	// The plugin helper is not valid for external plugins right now
	return run.WithVersion(name, version, updater)(adb, p.Meta, nil)
}
