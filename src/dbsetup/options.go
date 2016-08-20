package dbsetup

import (
	"config"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Options represents the configuration used for dbsetup. It is the source of
// truth used when creating/starting/stopping the database
type Options struct {
	// The directory in which ConnectorDB is set up
	DatabaseDirectory string `json:"database_directory"`

	// Config is used for Create.
	Config *config.Configuration `json:"config"`

	//These are optional - if they are set, an initial user is created on Create()
	//They are used only when passing a Configuration object to Create()
	InitialUser *config.UserMaker `json:"initial_user"`

	// The enabled options specify which services to set up/run
	FrontendEnabled bool `json:"frontend_running"`
	RedisEnabled    bool `json:"redis_running"`
	GnatsdEnabled   bool `json:"gnatsd_running"`
	SQLEnabled      bool `json:"sql_running"`

	// FrontendFlags are flags to pass to the frontend when starting it.
	// This allows using connectordb start with connectordb run's flags
	FrontendFlags []string
	FrontendPort  uint16 // Overload port
}

// Save saves the configuration
func (o *Options) Save(filename string) error {
	b, err := json.MarshalIndent(o, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0666)
}

// LoadOptions loads the options from a file
func LoadOptions(filename string) (*Options, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to load setup options from '%s': %s", filename, err.Error())
	}

	o := &Options{}
	err = json.Unmarshal(file, o)
	if err != nil {
		return nil, fmt.Errorf("Failed to load options from '%s': %s", filename, err.Error())
	}
	return o, nil
}
