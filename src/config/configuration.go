package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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

// GetSqlConnectionString returns the string used to connect to postgres
func (s *Service) GetSqlConnectionString() string {
	return fmt.Sprintf("postgres://%v:%v/connectordb?sslmode=disable", s.Hostname, s.Port)
}

//Configuration corresponds to the overall settings of ConnectorBD
type Configuration struct {
	Version int `json:"version"` //The database version that the configuration uses

	Service //This represents the overall connectordb frontend hostname/port

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

	SessionAuthKey       string `json:"session_authkey"`       //The key used to sign sessions
	SessionEncryptionKey string `json:"session_encryptionkey"` //The key used to encrypt sessions in cookies
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

		DisallowedNames: []string{"support", "www", "api"},

		//The defaults to use for the batch and chunks
		BatchSize: 250,
		ChunkSize: 5,

		SessionAuthKey:       base64.StdEncoding.EncodeToString(sessionAuthkey),
		SessionEncryptionKey: base64.StdEncoding.EncodeToString(sessionEncKey),
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
	return c, err
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
	if c.SessionAuthKey == "" {
		return securecookie.GenerateRandomKey(64), nil
	}

	return base64.StdEncoding.DecodeString(c.SessionAuthKey)
}

//GetSessionEncryptionKey returns the bytes associated with the config string
func (c *Configuration) GetSessionEncryptionKey() ([]byte, error) {
	//If no session encryption key is in config, generate one
	if c.SessionEncryptionKey == "" {
		return securecookie.GenerateRandomKey(32), nil
	}
	return base64.StdEncoding.DecodeString(c.SessionEncryptionKey)
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
