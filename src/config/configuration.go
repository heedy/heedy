/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/nu7hatch/gouuid"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"

	psconfig "github.com/connectordb/pipescript/config"

	log "github.com/Sirupsen/logrus"
)

//SqlType is the type of sql database used
const SqlType = "postgres"

// The header that is written to all config files
var configHeader = `/* ConnectorDB Configuration File

To see an explanation of the configuration options, please see:
https://github.com/connectordb/connectordb/blob/master/src/config/configuration.go
	Look at NewConfiguration() which explains defaults.

Particular configuration options:
frontend options: https://github.com/connectordb/connectordb/blob/master/src/config/frontend.go
	These are the options that pertain to the ConnectorDB server (REST API, web, request logging)
permissions: https://github.com/connectordb/connectordb/blob/master/src/config/permissions.go
	The permissions and access levels for each user type. All user types in the database are required.
access_levels: https://github.com/connectordb/connectordb/blob/master/src/config/accesslevel.go
	Specific access levels, which specify detailed read/write permissions

The configuration file supports javascript style comments. Comments are not inserted by default in this version
of ConnectorDB, because the JSON is generated automatically (it includes several custom values, such as auto-generated keys)

Several options support live reload. Changing them in the configuration file will automatically update the corresponding setting
in ConnectorDB. The ones that are not live-reloadable will not be reloaded (changing these options will not give any message).

When running a local database, the configuration file is in connectordb.pid in the database directory. It will be deleted on shutdown,
so will not save your changes. Save long-term changes to connectordb.conf in the same directory.
*/
`

// Configuration represents the options which are kept in a config file
type Configuration struct {
	Version int `json:"version"` // The version of the configuration file

	// Options pertaining to the frontend server.
	// These are transparent to json, so they appear directly in the main json.
	Frontend

	// Configuration options for a service
	Redis Service `json:"redis"`
	Nats  Service `json:"nats"`
	Sql   Service `json:"sql"`

	// The size of batches and chunks to use with the database
	BatchSize int `json:"batchsize"` // BatchSize is the number of datapoints per database entry
	ChunkSize int `json:"chunksize"` // ChunkSize is number of batches per database insert transaction

	// The cache sizes for users/devices/streams
	UseCache        bool  `json:"cache"` // Whether or not to enable caching
	UserCacheSize   int64 `json:"user_cache_size"`
	DeviceCacheSize int64 `json:"device_cache_size"`
	StreamCacheSize int64 `json:"stream_cache_size"`

	//These are optional - if they are set, an initial user is created on Create()
	//They are used only when passing a Configuration object to Create()
	InitialUsername        string `json:"createuser_username,omitempty"`
	InitialUserPassword    string `json:"createuser_password,omitempty"`
	InitialUserEmail       string `json:"createuser_email,omitempty"`
	InitialUserPermissions string `json:"createuser_permissions,omitempty"`

	// The prime number to use for scrambling IDs in the database.
	// WARNING: This must be CONSTANT! It should NEVER change after creating the database
	// http://preshing.com/20121224/how-to-generate-a-sequence-of-unique-random-integers/
	IDScramblePrime int64 `json:"database_id_scramble_prime"`

	// The given usernames are forbidden.
	DisallowedNames []string `json:"disallow_names"` //The names that are not permitted

	// The email suffixes that are permitted during user creation
	AllowedEmailSuffixes []string `json:"allowed_email_suffixes"`

	// The maximum number of users to allow. 0 means don't allow any new users, and -1 means unlimited
	// number of users
	MaxUsers int `json:"max_users"`

	// The configuration options for pipescript (https://github.com/connectordb/pipescript)
	PipeScript *psconfig.Configuration `json:"pipescript"`

	// The specific permissions granted to different user types
	// The required types are nobody and user.
	// If a user in the database has an unknown type, an error will be printed, and the user will fall back to
	// 'user' permissions (which are required)
	Permissions map[string]*Permissions `json:"permissions"`

	AccessLevels map[string]*AccessLevel `json:"access_levels"`

	// The following are exported fields that are used internally, and are not available to json.
	// This is honestly just lazy programming on my part - I am using the config struct as a temporary variable
	// placeholder when creating/starting a database... So technically it is part of the configuration, but it is
	// given explicitly as part of the command line args
	DatabaseDirectory string `json:"-"`
}

// NewConfiguration generates a configuration with reasonable defaults for use in ConnectorDB
func NewConfiguration() *Configuration {
	redispassword, _ := uuid.NewV4()
	natspassword, _ := uuid.NewV4()

	sessionAuthKey := securecookie.GenerateRandomKey(64)
	sessionEncKey := securecookie.GenerateRandomKey(32)

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
			//TODO: Have SQL accedd be auth'd
			Enabled: true,
		},

		Frontend: Frontend{
			Hostname: "0.0.0.0", // Host on all interfaces by default
			Port:     8000,

			Enabled: true,

			// Sets up the session cookie keys that are used
			CookieSession: CookieSession{
				AuthKey:       base64.StdEncoding.EncodeToString(sessionAuthKey),
				EncryptionKey: base64.StdEncoding.EncodeToString(sessionEncKey),
				MaxAge:        60 * 60 * 24 * 30 * 4, //About 4 months is the default expiration time of a cookie
			},

			// By default, captcha is disabled
			Captcha: Captcha{
				Enabled: false,
			},

			// By default log query counts once a minute, and display server statistics
			// once a day
			QueryDisplayTimer: 60,
			StatsDisplayTimer: 60 * 60 * 24,

			// Why not minify? Turning it off is useful for debugging - but users outnumber coders by a large margin.
			Minify: true,
		},

		//The defaults to use for the batch and chunks
		BatchSize: 250,
		ChunkSize: 5,

		UseCache:        true,
		UserCacheSize:   1000,
		DeviceCacheSize: 10000,
		StreamCacheSize: 10000,

		// This is the CONSTANT default. The database will explode if this is ever changed.
		// You have been warned.
		IDScramblePrime: 2147483423,

		// Disallowed names are names that would conflict with the ConnectorDB frontend
		DisallowedNames: []string{"support", "www", "api", "app", "favicon.ico", "robots.txt", "sitemap.xml", "join", "login", "user", "admin", "nobody", "root"},

		// Allow an arbitrary number of users by default
		MaxUsers: -1,

		// Use the default settings.
		// NOTE: Once a configuration is loaded,
		PipeScript: psconfig.Default(),

		Permissions: map[string]*Permissions{
			"nobody": &Permissions{
				Join:                false,
				JoinDisabledMessage: "You must be logged in as admin to add users",

				PublicReadAccessLevel:  "none",
				PrivateReadAccessLevel: "none",
				SelfReadAccessLevel:    "none",

				PublicWriteAccessLevel:  "none",
				PrivateWriteAccessLevel: "none",
				SelfWriteAccessLevel:    "none",

				OwnerDeviceReadAccessLevel:  "none",
				OwnerDeviceWriteAccessLevel: "none",
			},
			"user": &Permissions{
				Join:                false,
				JoinDisabledMessage: "You must be logged in as admin to add users",

				PublicReadAccessLevel:  "publicread",
				PrivateReadAccessLevel: "none",
				SelfReadAccessLevel:    "selfread",

				PublicWriteAccessLevel:  "none",
				PrivateWriteAccessLevel: "none",
				SelfWriteAccessLevel:    "selfwrite",

				OwnerDeviceReadAccessLevel:  "selfread",
				OwnerDeviceWriteAccessLevel: "selfwrite",
			},
			"admin": &Permissions{
				Join:                true,
				JoinDisabledMessage: "Join is disabled",

				PublicReadAccessLevel:  "full",
				PrivateReadAccessLevel: "full",
				SelfReadAccessLevel:    "full",

				PublicWriteAccessLevel:  "full",
				PrivateWriteAccessLevel: "full",
				SelfWriteAccessLevel:    "full",

				OwnerDeviceReadAccessLevel:  "full",
				OwnerDeviceWriteAccessLevel: "full",
			},
		},

		AccessLevels: map[string]*AccessLevel{
			"publicread": &AccessLevel{
				CanAccessUser:                   true,
				CanAccessDevice:                 true,
				CanAccessStream:                 true,
				CanAccessNonUserEditableDevices: false,
				UserName:                        true,
				UserNickname:                    true,
				UserEmail:                       true,
				UserDescription:                 true,
				UserIcon:                        true,
				UserPermissions:                 false,
				UserPublic:                      true,
				UserPassword:                    false,
				DeviceName:                      true,
				DeviceNickname:                  true,
				DeviceDescription:               true,
				DeviceIcon:                      true,
				DeviceAPIKey:                    false,
				DeviceEnabled:                   true,
				DeviceIsVisible:                 true,
				DeviceUserEditable:              true,
				DevicePublic:                    true,
				DeviceCanReadUser:               false,
				DeviceCanReadExternal:           false,
				DeviceCanWriteUser:              false,
				DeviceCanWriteExternal:          false,
				StreamName:                      true,
				StreamNickname:                  true,
				StreamDescription:               true,
				StreamIcon:                      true,
				StreamSchema:                    true,
				StreamEphemeral:                 true,
				StreamDownlink:                  true,
			},
			"selfwrite": &AccessLevel{
				CanAccessUser:                   true,
				CanAccessDevice:                 true,
				CanAccessStream:                 true,
				CanAccessNonUserEditableDevices: false,
				UserName:                        false,
				UserNickname:                    true,
				UserEmail:                       true,
				UserDescription:                 true,
				UserIcon:                        true,
				UserPermissions:                 false,
				UserPublic:                      true,
				UserPassword:                    true,
				DeviceName:                      false,
				DeviceNickname:                  true,
				DeviceDescription:               true,
				DeviceIcon:                      true,
				DeviceAPIKey:                    true,
				DeviceEnabled:                   true,
				DeviceIsVisible:                 true,
				DeviceUserEditable:              false,
				DevicePublic:                    true,
				DeviceCanReadUser:               true,
				DeviceCanReadExternal:           true,
				DeviceCanWriteUser:              true,
				DeviceCanWriteExternal:          true,
				StreamName:                      false,
				StreamNickname:                  true,
				StreamDescription:               true,
				StreamIcon:                      true,
				StreamSchema:                    true,
				StreamEphemeral:                 true,
				StreamDownlink:                  true,
			},
			"selfread": &AccessLevel{
				CanAccessUser:                   true,
				CanAccessDevice:                 true,
				CanAccessStream:                 true,
				CanAccessNonUserEditableDevices: true,
				UserName:                        true,
				UserNickname:                    true,
				UserEmail:                       true,
				UserDescription:                 true,
				UserIcon:                        true,
				UserPermissions:                 true,
				UserPublic:                      true,
				UserPassword:                    false,
				DeviceName:                      true,
				DeviceNickname:                  true,
				DeviceDescription:               true,
				DeviceIcon:                      true,
				DeviceAPIKey:                    true,
				DeviceEnabled:                   true,
				DeviceIsVisible:                 true,
				DeviceUserEditable:              true,
				DevicePublic:                    true,
				DeviceCanReadUser:               true,
				DeviceCanReadExternal:           true,
				DeviceCanWriteUser:              true,
				DeviceCanWriteExternal:          true,
				StreamName:                      true,
				StreamNickname:                  true,
				StreamDescription:               true,
				StreamIcon:                      true,
				StreamSchema:                    true,
				StreamEphemeral:                 true,
				StreamDownlink:                  true,
			},
		},
	}

}

// GetAccessLevel returns the given access level
func (c *Configuration) GetAccessLevel(level string) (*AccessLevel, error) {
	if level == "none" {
		return &NoneAccessLevel, nil
	}
	if level == "full" {
		return &FullAccessLevel, nil
	}
	al, ok := c.AccessLevels[level]
	if !ok {
		return nil, fmt.Errorf("Could not find access level '%s'", level)
	}
	return al, nil
}

// GetSqlConnectionString returns the string used to connect to postgres
func (c *Configuration) GetSqlConnectionString() string {
	return c.Sql.GetSqlConnectionString()
}

// IsAllowedUsername checks if the user name is allowed by the configuration
func (c *Configuration) IsAllowedUsername(name string) bool {
	for i := range c.DisallowedNames {
		if name == c.DisallowedNames[i] {
			return false
		}
	}
	return true
}

// IsAllowedEmail checks in the configuration's AllowedEmailSuffixes to see if
// the given email address is valid allowed to sign up.
func (c *Configuration) IsAllowedEmail(emailAddress string) bool {
	if len(c.AllowedEmailSuffixes) == 0 {
		return true
	}

	for _, suffix := range c.AllowedEmailSuffixes {
		if strings.HasSuffix(emailAddress, suffix) {
			return true
		}
	}

	return false
}

// String returns a string representation of the configuration
func (c *Configuration) String() string {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return string(b)
}

// Save saves the configuration
func (c *Configuration) Save(filename string) error {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(configHeader))
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// Load a configuration from the given file, and ensures that it is valid
func Load(filename string) (*Configuration, error) {
	log.Debugf("Loading configuration from %s", filename)

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// To allow comments in the json, we minify the file with js minifer before parsing
	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)
	file, err = m.Bytes("text/javascript", file)
	if err != nil {
		return nil, err
	}

	c := NewConfiguration()
	err = json.Unmarshal(file, c)
	if err != nil {
		return nil, err
	}

	// Before doing anything, we need to change the working directory to that of the config file.
	// We switch back to the current working dir once done validating.
	// Validation takes any file names and converts them to absolute paths.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}
	// Change the directory back on exit
	defer os.Chdir(cwd)

	return c, c.Validate()
}
