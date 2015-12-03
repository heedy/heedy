package config

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/securecookie"
	"github.com/nu7hatch/gouuid"
)

//SqlType is the type of sql database used
const SqlType = "postgres"

//Service represents a single server connection
type Service struct {
	Hostname string `json:"hostname"`
	Port     uint16 `json:"port"`

	Username string `json:"username,omitempty"` //Username and password are used for login to constituent servers
	Password string `json:"password,omitempty"`

	//SSLPort uint16 `json:"sslport"` //The port on which to run Stunnel

	Enabled bool `json:"enabled"` //Whether or not to run the service on "start"
}

//ConnectionString returns the connection string
func (s *Service) ConnectionString() string {
	return fmt.Sprintf("%s:%v", s.Hostname, s.Port)
}

// GetSqlConnectionString returns the string used to connect to postgres
func (s *Service) GetSqlConnectionString() string {
	return fmt.Sprintf("postgres://%v:%v/connectordb?sslmode=disable", s.Hostname, s.Port)
}

//Session refers to a cookie session
type Session struct {
	AuthKey       string `json:"authkey"`       //The key used to sign sessions
	EncryptionKey string `json:"encryptionkey"` //The key used to encrypt sessions in cookies
	MaxAge        int    `json:"maxage"`        //The maximum age of a cookie in a session (seconds)
}

//Configuration corresponds to the overall settings of ConnectorBD
type Configuration struct {
	Version int `json:"version"` //The database version that the configuration uses

	Service //This represents the overall connectordb frontend hostname/port

	//These two options enable https on the server
	TLSKey  string `json:"tls_key"`
	TLSCert string `json:"tls_cert"`

	Redis Service `json:"redis"`
	Nats  Service `json:"nats"`
	Sql   Service `json:"sql"`

	//BDWriter specifies whether to run the DBWriter on this config when starting
	//DBWriter bool `json:"dbwriter"`

	DatabaseDirectory string `json:"-"`

	DisallowedNames []string `json:"disallow_names"` //The names that are not permitted

	//These are optional - if they are set, an initial user is created on Create()
	//They are used only when passing a Configuration object to Create()
	InitialUsername     string `json:"-"`
	InitialUserPassword string `json:"-"`
	InitialUserEmail    string `json:"-"`

	//The size of batches and chunks to use with the database
	BatchSize int `json:"batchsize"`
	ChunkSize int `json:"chunksize"`

	//The following options are for the ConnectorDB server
	AllowJoin bool `json:"allow_join"` //Whether or not to permit adding of users through web interface

	Session Session `json:"session"` //The session cookies to allow in the website

	SiteName string `json:"sitename"` //The site to use for requests and stuff

	AllowCrossOrigin bool `json:"allowcrossorigin"` //Whether the site options permit CORS
}

//NewConfiguration generates a configuration for the database.
func NewConfiguration() *Configuration {
	redispassword, _ := uuid.NewV4()
	natspassword, _ := uuid.NewV4()

	sessionAuthkey := securecookie.GenerateRandomKey(64)
	sessionEncKey := securecookie.GenerateRandomKey(32)

	//sqlpassword, _ := uuid.NewV4()
	return &Configuration{
		Version: 1,
		Redis: Service{
			Hostname: "localhost",
			Port:     6379,
			Password: redispassword.String(),
			Enabled:  true,
		},
		Nats: Service{
			Hostname: "localhost",
			Port:     4222,
			Username: "connectordb",
			Password: natspassword.String(),
			Enabled:  true,
		},
		Sql: Service{
			Hostname: "localhost",
			Port:     52592,
			//Username: "connectordb",
			//Password: sqlpassword,
			Enabled: true,
		},
		//DBWriter: true,

		//The ConnectorDB frontend server
		Service: Service{
			Hostname: "0.0.0.0",
			Port:     8000,
			Enabled:  false,
		},

		DisallowedNames: []string{"support", "www", "api", "app", "favicon.ico", "robots.txt"},

		//The defaults to use for the batch and chunks
		BatchSize: 250,
		ChunkSize: 5,

		Session: Session{
			AuthKey:       base64.StdEncoding.EncodeToString(sessionAuthkey),
			EncryptionKey: base64.StdEncoding.EncodeToString(sessionEncKey),
			MaxAge:        60 * 60 * 24 * 30 * 4, //About 4 months is the default expiration time of a cookie
		},

		SiteName: "",

		AllowCrossOrigin: false,
	}
}

//Load gets a configuration from a file
func Load(filename string) (c *Configuration, err error) {
	log.Debugf("Loading configuration from %s", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c = NewConfiguration()
	err = json.Unmarshal(file, c)
	if err != nil {
		return nil, err
	}

	//Before loading files from the configuration, we must change the cwd to the config directory, and then change it back to the current one
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}

	err = c.InitMissing()

	//Now we move back to the current working directory
	err = os.Chdir(cwd)
	if err != nil {
		return nil, err
	}

	return c, err
}

// TLSEnabled returns whether the server is to run in https mode
func (c *Configuration) TLSEnabled() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}

// InitMissing Sets up missing values with reasonable defaults
func (c *Configuration) InitMissing() error {
	if c.Hostname == "" {
		c.Hostname = "localhost"
	}

	if c.SiteName == "" {
		//Assume we are testing: set the sitename to localhost
		if (c.Port == 80 && !c.TLSEnabled()) || (c.Port == 443 && c.TLSEnabled()) {
			//No need to include port number
			c.SiteName = c.Hostname
		} else {
			c.SiteName = c.ConnectionString()
		}
	}
	if !strings.HasPrefix(c.SiteName, "http://") && !strings.HasPrefix(c.SiteName, "https://") {
		//Site name does not start with http - so we choose it based upon whether tls is enabled
		if c.TLSEnabled() {
			c.SiteName = "https://" + c.SiteName
		} else {
			c.SiteName = "http://" + c.SiteName
		}
	}

	if c.BatchSize <= 0 {
		c.BatchSize = 250
	}
	if c.ChunkSize <= 0 {
		c.ChunkSize = 5
	}

	if c.Session.AuthKey == "" {
		sessionAuthkey := securecookie.GenerateRandomKey(64)
		c.Session.AuthKey = base64.StdEncoding.EncodeToString(sessionAuthkey)
	}
	if c.Session.EncryptionKey == "" {
		sessionEncKey := securecookie.GenerateRandomKey(32)
		c.Session.EncryptionKey = base64.StdEncoding.EncodeToString(sessionEncKey)
	}
	if c.Session.MaxAge <= 0 {
		c.Session.MaxAge = 60 * 60 * 24 * 30 * 4
	}
	if len(c.DisallowedNames) == 0 {
		c.DisallowedNames = []string{"support", "www", "api", "app", "favicon.ico", "robots.txt"}
	}

	//Make sure the TLS cert and key are valid
	if c.TLSCert != "" || c.TLSKey != "" {
		log.Debug("Checking TLS Keys")
		_, err := tls.LoadX509KeyPair(c.TLSCert, c.TLSKey)
		if err != nil {
			return err
		}

		//Set the file paths to be full paths
		c.TLSCert, err = filepath.Abs(c.TLSCert)
		if err != nil {
			return err
		}
		c.TLSKey, err = filepath.Abs(c.TLSKey)
		if err != nil {
			return err
		}
	}

	return nil
}

//String returns a string representation of the configuration
func (c *Configuration) String() string {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return string(b)
}

//Save saves the configuration
func (c *Configuration) Save(filepath string) error {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, b, os.FileMode(0755))
}

//GetSessionAuthKey returns the bytes associated with the config string
func (c *Configuration) GetSessionAuthKey() ([]byte, error) {
	//If no session key is in config, generate one
	if c.Session.AuthKey == "" {
		return securecookie.GenerateRandomKey(64), nil
	}

	return base64.StdEncoding.DecodeString(c.Session.AuthKey)
}

//GetSessionEncryptionKey returns the bytes associated with the config string
func (c *Configuration) GetSessionEncryptionKey() ([]byte, error) {
	//If no session encryption key is in config, generate one
	if c.Session.EncryptionKey == "" {
		return securecookie.GenerateRandomKey(32), nil
	}
	return base64.StdEncoding.DecodeString(c.Session.EncryptionKey)
}

//GetSqlConnectionString Returns the database connection string for the current database
func (c *Configuration) GetSqlConnectionString() string {
	return c.Sql.GetSqlConnectionString()
}

//Options generates the ConnectorDB optinos based upon the given configuration
func (c *Configuration) Options() *Options {
	opt := NewOptions()

	opt.RedisOptions.Addr = fmt.Sprintf("%s:%d", c.Redis.Hostname, c.Redis.Port)
	opt.RedisOptions.Password = c.Redis.Password

	opt.NatsOptions.Url = fmt.Sprintf("nats://%s:%s@%s:%d", c.Nats.Username, c.Nats.Password, c.Nats.Hostname, c.Nats.Port)

	opt.SqlConnectionString = c.GetSqlConnectionString()

	opt.BatchSize = c.BatchSize
	opt.ChunkSize = c.ChunkSize

	return opt
}

// GetRedisURI returns the redis "uri", no prefix appended
func (c *Configuration) GetRedisURI() string {
	return fmt.Sprintf("%s:%d", c.Redis.Hostname, c.Redis.Port)
}

// GetGnatsdURI gets the gnatsd "uri" no prefix appended; it'll be in the format host:port
func (c *Configuration) GetGnatsdURI() string {
	return fmt.Sprintf("%s:%d", c.Nats.Hostname, c.Nats.Port)
}
