package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

//Configuration corresponds to the overall settings of ConnectorBD
type Configuration struct {
	RedisPort int
	RedisHost string

	GnatsdPort int
	GnatsdHost string

	PostgresPort int
	PostgresHost string

	//These are optional - if they are set, an initial user is created on Create()
	Username     string `json:"-"`
	UserPassword string `json:"-"`
	UserEmail    string `json:"-"`

	DatabaseDirectory string `json:"-"`

	Host string //The host at which to serve
	Port int    //The port on which to run the server

	DisallowedNames []string //The names that are not permitted

}

//NewConfiguration generates a configuration for the database.
func NewConfiguration() *Configuration {
	return &Configuration{
		RedisPort:    6379,
		RedisHost:    "localhost",
		GnatsdPort:   4222,
		GnatsdHost:   "localhost",
		PostgresHost: "localhost",
		PostgresPort: 52592,
		Port:         8000,

		DisallowedNames: []string{"support", "www", "api"},
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

//GetDatabaseConnectionString Returns the database connection string for the current database
func (c *Configuration) GetDatabaseConnectionString() string {
	return fmt.Sprintf("postgres://%v:%v/connectordb?sslmode=disable", c.PostgresHost, c.PostgresPort)
}

//Options generates the ConnectorDB optinos based upon the given configuration
func (c *Configuration) Options() *Options {
	opt := NewOptions()

	opt.RedisOptions.Addr = fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
	opt.NatsOptions.Url = fmt.Sprintf("nats://%s:%d", c.GnatsdHost, c.GnatsdPort)

	opt.SqlConnectionString = c.GetDatabaseConnectionString()

	return opt
}

// Returns the redis "uri", no prefix appneded
func (c *Configuration) GetRedisUri() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

// Get the gnatsd "uri" no prefix appended; it'll be in the format host:port
func (c *Configuration) GetGnatsdUri() string {
	return fmt.Sprintf("%s:%d", c.GnatsdHost, c.GnatsdPort)
}
