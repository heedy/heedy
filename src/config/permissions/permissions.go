package permissions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"

	log "github.com/Sirupsen/logrus"
)

// The header to write at the top of the permissions file
var permissionsHeader = `/*
ConnectorDB Permissions Specification.

Permissions files enable you to specify detailed user types, limits, roles,
and allowed actions for ConnectorDB. To load this permissions file, put
its location into the 'permissions' field of your ConnectorDB configuration file.
Note that ConnectorDB watches the loaded permissions file for changes, and automatically
updates the internal permissions on changes to the file.

To generate the default permissions into a "permissions.conf" file, run:
connectordb permissions permissions.conf

This will generate the "default" configuration (which is built-in).
For more info, see:
github.com/connectordb/connectordb/tree/master/src/config/permissions/default.go

Please note that "full" and "none" are built-in access levels, and as such
are not shown in the file.
*/
`

// Permissions is the structure which represents all things that are allowed and disallowed for all
// types of users and visitors to a ConnectorDB powered server
type Permissions struct {
	Version int  `json:"version"`
	Watch   bool `json:"watch"` // Whether or not to watch the file for changes

	// The given usernames are forbidden.
	DisallowedNames []string `json:"disallow_names"` //The names that are not permitted

	// The email suffixes that are permitted during user creation
	AllowedEmailSuffixes []string `json:"allowed_email_suffixes"`

	// The maximum number of users to allow. 0 means don't allow any new users, and -1 means unlimited
	// number of users
	MaxUsers int `json:"max_users"`

	// The specific permissions granted to different user types
	// The required types are nobody and user.
	// If a user in the database has an unknown role, an error will be printed, and the user will fall back to
	// 'nobody' role.
	Roles map[string]*Role `json:"roles"`

	AccessLevels map[string]*AccessLevel `json:"access_levels"`
}

// GetAccessLevel returns the given access level
func (p *Permissions) GetAccessLevel(level string) (*AccessLevel, error) {
	if level == "none" {
		return &NoneAccessLevel, nil
	}
	if level == "full" {
		return &FullAccessLevel, nil
	}
	al, ok := p.AccessLevels[level]
	if !ok {
		return nil, fmt.Errorf("Could not find access level '%s'", level)
	}
	return al, nil
}

// IsAllowedUsername checks if the user name is allowed by the permissions
func (p *Permissions) IsAllowedUsername(name string) bool {
	for i := range p.DisallowedNames {
		if name == p.DisallowedNames[i] {
			return false
		}
	}
	return true
}

// IsAllowedEmail checks in the permissions AllowedEmailSuffixes to see if
// the given email address is valid allowed to sign up.
func (p *Permissions) IsAllowedEmail(emailAddress string) bool {
	if len(p.AllowedEmailSuffixes) == 0 {
		return true
	}

	for _, suffix := range p.AllowedEmailSuffixes {
		if strings.HasSuffix(emailAddress, suffix) {
			return true
		}
	}

	return false
}

// String returns a json representation of the Permissions object
func (p *Permissions) String() string {
	b, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return string(b)
}

// Save writes the permissions to the given file name
func (p *Permissions) Save(filename string) error {
	b, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(permissionsHeader))
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// Load loads permissions from the given file name, giving an error if loading fails
// or if the permissions do not pass validation
func Load(filename string) (*Permissions, error) {
	if filename == "default" {
		return &Default, nil
	}
	log.Debugf("Loading Permissions from %s", filename)

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

	p := &Permissions{}
	err = json.Unmarshal(file, p)
	if err != nil {
		return nil, err
	}
	return p, p.Validate()
}
