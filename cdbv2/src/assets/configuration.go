package assets

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// Configuration represents the options that are loaded from a config file
type Configuration struct {
	Hostname string `toml:"hostname"`
	Port     uint16 `toml:"port"`
	SiteURL  string `toml:"siteurl"`
}

// NewConfiguration returns a new empty configuration
func NewConfiguration() {
	return &Configuration{}
}

// LoadFile loads configuration from a file name, passing it to LoadString
func (c *Configuration) LoadFile(string filename) error {
	log.Debugf("Merging in configuration from %s", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return c.Load(data)
}

// Load updates the configuration from a string configuration file.
// Keeps any values that were not overloaded at their
// current values
func (c *Configuration) Load(string filecontents) error {
	_, err := toml.Decode(fileContents, c)
	return err
}

// Validate ensures that the configuration does not have messed up values
func (c *Configuration) Validate() error {
	// TODO: Implement validation
	return nil
}
