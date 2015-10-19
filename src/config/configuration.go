package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
)

//Service represents a single server connection
type Service struct {
	Hostname string `json:"hostname"`
	Port     uint16 `json:"port"`

	Username string `json:"username,omitempty"` //Username and password are used for login to constituent servers
	Password string `json:"password,omitempty"`

	//SSLPort uint16 `json:"sslport"` //The port on which to run Stunnel

	Enabled bool `json:"enabled"` //Whether or not to run the service on "start"
}

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
	DBWriter bool `json:"dbwriter"`

	DatabaseDirectory string `json:"-"`

	DisallowedNames []string `json:"disallow_names"` //The names that are not permitted

	//These are optional - if they are set, an initial user is created on Create()
	//They are used only when passing a Configuration object to Create()
	InitialUsername     string `json:"-"`
	InitialUserPassword string `json:"-"`
	InitialUserEmail    string `json:"-"`
}

//NewConfiguration generates a configuration for the database.
func NewConfiguration() *Configuration {
	redispassword, _ := uuid.NewV4()
	natspassword, _ := uuid.NewV4()
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
		DBWriter: true,

		//The ConnectorDB frontend server
		Service: Service{
			Hostname: "0.0.0.0",
			Port:     8000,
			Enabled:  false,
		},

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

	return opt
}

// Returns the redis "uri", no prefix appneded
func (c *Configuration) GetRedisUri() string {
	return fmt.Sprintf("%s:%d", c.Redis.Hostname, c.Redis.Port)
}

// Get the gnatsd "uri" no prefix appended; it'll be in the format host:port
func (c *Configuration) GetGnatsdUri() string {
	return fmt.Sprintf("%s:%d", c.Nats.Hostname, c.Nats.Port)
}
