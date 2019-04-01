package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Details is used in groups, users, connections and streams to hold info
type Details struct {
	ID          string  `json:"id"`
	Name        *string `json:"name"`
	FullName    *string `json:"fullname"`
	Description *string `json:"description"`
	Owner       *string `json:"owner"`
	Avatar      *string `json:"avatar"`
}

// Group holds a group's details
type Group struct {
	Details

	PublicAccess *int `json:"public_access,omitempty" db:"public_access"`
	UserAccess   *int `json:"user_access,omitempty" db:"user_access"`
}

// User holds a user's data
type User struct {
	Group
	Password string `json:"password,omitempty"`
}

type Connection struct {
	Details

	APIKey *string `json:"apikey,omitempty"`

	Settings      *string `json:"settings,omitempty"`
	SettingSchema *string `json:"setting_schema,omitempty" db:"setting_schema"`
}

type Stream struct {
	Details

	Connection *string `json:"connection,omitempty"`
	Schema     *string `json:"schema,omitempty"`
	External   *string `json:"external,omitempty"`
	Actor      *bool   `json:"actor,omitempty"`
	Access     *int    `json:"access,omitempty"`
}

// DB represents the database. This interface is implemented in many ways:
//	once for admin
//	once for users
//	once for connections
//	once for public
type DB interface {
	AdminDB() *AdminDB // Returns the underlying administrative database
	ID() string        // This is an identifier for the database. empty string is public access

	// Currently logged in user
	User() (*User, error)

	CreateUser(u *User) error
	ReadUser(name string, avatar bool) (*User, error)
	UpdateUser(u *User) error
	DelUser(name string) error

	GetUserScopes(name string) ([]string, error)
}

var (
	ErrNotFound        = errors.New("not_found: The selected resource was not found")
	ErrNoUpdate        = errors.New("Nothing to update")
	ErrNoPasswordGiven = errors.New("A user cannot have an empty password")
	ErrUserNotFound    = errors.New("User was not found")
	ErrInvalidName     = errors.New("Invalid name")
	ErrInvalidQuery    = errors.New("Invalid query")
	ErrAccessDenied    = errors.New("access_denied: you are not allowed to do this")
)

// GenerateKey creates a random API key
func GenerateKey(length int) (string, error) {
	// Prepare the plugin API key
	apikey := make([]byte, length)
	_, err := rand.Read(apikey)
	return base64.StdEncoding.EncodeToString(apikey), err
}

// HashPassword generates a bcrypt hash for the given password
func HashPassword(password string) (string, error) {
	passwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(passwd), err
}

// CheckPassword checks if the password is valid
func CheckPassword(password, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

var (
	nameValidator = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_-]*$")
)

func ValidName(name string) error {
	if nameValidator.MatchString(name) && len(name) > 0 {
		return nil
	}
	return ErrInvalidName
}

// Ensures that the avatar is in a valid format
func ValidAvatar(avatar string) error {
	if avatar == "" {
		return nil
	}
	// We permit special avatar prefixes to be used. The first one is material:, which represents material icons
	// that are assumed to be bundled with all applications that display heedy data. The second is fa: which
	// will represent fontawesome avatars in the future
	if strings.HasPrefix(avatar, "material:") || strings.HasPrefix(avatar, "fa:") {
		if len(avatar) > 30 {
			return errors.New("bad_request: avatar icon name can't be more than 30 characters")
		}
		return nil
	}
	if !strings.HasPrefix(avatar, "data:image/") {
		return errors.New("bad_request: Avatar iamges must be data-urls")
	}
	return nil
}

// Performs a set of tests on the result and error of a
// call to see what kind of error we should return.
func getExecError(result sql.Result, err error) error {
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func extractDetails(d *Details) (detailColumns []string, detailValues []interface{}, err error) {
	detailColumns = make([]string, 0)
	detailValues = make([]interface{}, 0)

	if d.Name != nil {
		if err = ValidName(*d.Name); err != nil {
			return
		}
		detailColumns = append(detailColumns, "name")
		detailValues = append(detailValues, *d.Name)
	}

	if d.Description != nil {
		detailColumns = append(detailColumns, "description")
		detailValues = append(detailValues, *d.Description)
	}
	if d.Avatar != nil {
		if err = ValidAvatar(*d.Avatar); err != nil {
			return
		}
		detailColumns = append(detailColumns, "avatar")
		detailValues = append(detailValues, *d.Avatar)
	}
	if d.FullName != nil {
		detailColumns = append(detailColumns, "fullname")
		detailValues = append(detailValues, *d.FullName)
	}
	if d.Owner != nil {
		if err = ValidName(*d.Owner); err != nil {
			return
		}
		detailColumns = append(detailColumns, "owner")
		detailValues = append(detailValues, *d.Owner)
	}
	return
}

// Checks whether the given group access level is OK
func ValidGroupAccessLevel(level int) error {
	if level < 0 || level > 600 {
		return errors.New("malformed_query: Access levels are 0-600")
	}
	return nil
}

func extractGroup(g *Group) (groupColumns []string, groupValues []interface{}, err error) {
	groupColumns, groupValues, err = extractDetails(&g.Details)

	if g.PublicAccess != nil {
		if err = ValidGroupAccessLevel(*g.PublicAccess); err != nil {
			return
		}
		groupColumns = append(groupColumns, "public_access")
		groupValues = append(groupValues, *g.PublicAccess)
	}
	if g.UserAccess != nil {
		if err = ValidGroupAccessLevel(*g.PublicAccess); err != nil {
			return
		}
		groupColumns = append(groupColumns, "user_access")
		groupValues = append(groupValues, *g.UserAccess)
	}
	return
}

func extractUser(u *User) (groupColumns []string, groupValues []interface{}, userColumns []string, userValues []interface{}, err error) {
	groupColumns, groupValues, err = extractGroup(&u.Group)
	if err != nil {
		return
	}

	userColumns = make([]string, 0)
	userValues = make([]interface{}, 0)

	if u.Password != "" {
		var password string
		password, err = HashPassword(u.Password)
		if err != nil {
			return
		}
		userColumns = append(userColumns, "password")
		userValues = append(userValues, password)
	}

	if len(userColumns) < 1 && len(groupColumns) < 1 {
		err = ErrNoUpdate
	}
	return
}

func extractConnection(c *Connection) (cColumns []string, cValues []interface{}, err error) {
	cColumns, cValues, err = extractDetails(&c.Details)
	if err != nil {
		return
	}
	if c.Settings != nil {
		cColumns = append(cColumns, "settings")
		cValues = append(cValues, *c.Settings)
	}
	if c.SettingSchema != nil {
		cColumns = append(cColumns, "setting_schema")
		cValues = append(cValues, *c.SettingSchema)
	}

	// Guaranteed to be last element
	if c.APIKey != nil {
		cColumns = append(cColumns, "apikey")
		if *c.APIKey == "" {
			// This means deleting the API key, so set it to empty
			cValues = append(cValues, "")
		} else {
			// Anything else we replace with a new API key
			var apikey string
			apikey, err = GenerateKey(15)
			if err != nil {
				return
			}
			c.APIKey = &apikey // Write the API key back to the connection object
			cValues = append(cValues, apikey)
		}

	}

	return
}

func extractStream(s *Stream) (sColumns []string, sValues []interface{}, err error) {
	sColumns, sValues, err = extractDetails(&s.Details)
	if err != nil {
		return
	}

	if s.Connection != nil {
		sColumns = append(sColumns, "connection")
		sValues = append(sValues, *s.Connection)
	}
	if s.Schema != nil {
		sColumns = append(sColumns, "schema")
		sValues = append(sValues, *s.Schema)
	}
	if s.External != nil {
		sColumns = append(sColumns, "external")
		sValues = append(sValues, *s.External)
	}
	if s.Actor != nil {
		sColumns = append(sColumns, "actor")
		sValues = append(sValues, *s.Actor)
	}

	if s.Access != nil {
		sColumns = append(sColumns, "access")
		sValues = append(sValues, *s.Access)
	}

	return
}

// Insert the right amount of question marks for the given query
func qQ(size int) string {
	s := strings.Repeat("?,", size)
	return s[:len(s)-1]
}

func userCreateQuery(u *User) (string, []interface{}, string, []interface{}, error) {
	if u.Name == nil {
		return "", nil, "", nil, ErrInvalidName
	}
	if u.Password == "" {
		return "", nil, "", nil, ErrNoPasswordGiven
	}
	if u.Owner != nil {
		return "", nil, "", nil, ErrInvalidQuery
	}
	u.Owner = u.Name

	groupColumns, groupValues, userColumns, userValues, err := extractUser(u)
	if err != nil {
		return "", nil, "", nil, err
	}

	// Now add the name of the user as the ID of the details (group), and as the name of the user
	groupColumns = append(groupColumns, "id")
	groupValues = append(groupValues, *u.Name)

	userColumns = append(userColumns, "name")
	userValues = append(userValues, u.Name)
	u.ID = *u.Name

	return strings.Join(groupColumns, ","), groupValues, strings.Join(userColumns, ","), userValues, err
}

func userUpdateQuery(u *User) (string, []interface{}, string, []interface{}, error) {
	if err := ValidName(u.ID); err != nil {
		return "", nil, "", nil, err
	}
	if u.Owner != nil {
		u.Owner = nil
	}
	gColumns, groupValues, uColumns, userValues, err := extractUser(u)
	if err != nil {
		return "", nil, "", nil, err
	}

	groupColumns := strings.Join(gColumns, "=?,") + "=?"
	userColumns := strings.Join(uColumns, "=?,") + "=?"

	if u.Name != nil {
		// A name change changes the group's ID also. We need to manually handle this.
		groupValues = append(groupValues, *u.Name)
		groupColumns = groupColumns + ",id=?"

		// Unfortunately, sqlite doesn't support alter table foreign keys, so we need to manually update the user name,
		// rather than cascading id change to user name
		userValues = append(userValues, *u.Name)
		if len(userValues) > 1 {
			userColumns = userColumns + ",name=?"
		} else {
			userColumns = "name=?"
		}
	}

	userValues = append(userValues, u.ID)
	groupValues = append(groupValues, u.ID)

	return groupColumns, groupValues, userColumns, userValues, err
}

func groupCreateQuery(g *Group) (string, []interface{}, error) {
	if g.Name == nil {
		return "", nil, ErrInvalidName
	}
	if g.Owner == nil {
		// A group must have an owner
		return "", nil, ErrInvalidQuery
	}
	groupColumns, groupValues, err := extractGroup(g)
	if err != nil {
		return "", nil, err
	}

	// Since we are creating the details, we also set up the id of the group
	// We guarantee that ID is last element
	groupColumns = append(groupColumns, "id")
	gid := uuid.New().String()
	groupValues = append(groupValues, gid)
	g.ID = gid // Set the object's ID

	return strings.Join(groupColumns, ","), groupValues, nil

}

func groupUpdateQuery(g *Group) (string, []interface{}, error) {
	groupColumns, groupValues, err := extractGroup(g)
	return strings.Join(groupColumns, "=?,") + "=?", groupValues, err
}

func connectionCreateQuery(c *Connection) (string, []interface{}, error) {
	if c.Name == nil {
		return "", nil, ErrInvalidName
	}
	if c.Owner == nil {
		return "", nil, ErrInvalidQuery
	}
	if c.APIKey == nil {
		// We want the API key to always be set on create - and set to null if none specified
		// This is because we use the cValues array in returns, so we want the last 2 elements
		// to be the API key, and id.
		es := ""
		c.APIKey = &es
	}
	cColumns, cValues, err := extractConnection(c)
	if err != nil {
		return "", nil, err
	}

	// We create an ID for the connection. Guaranteed to be last element
	cColumns = append(cColumns, "id")
	cid := uuid.New().String()
	cValues = append(cValues, cid)
	c.ID = cid

	return strings.Join(cColumns, ","), cValues, err
}

func connectionUpdateQuery(c *Connection) (string, []interface{}, error) {
	cColumns, cValues, err := extractConnection(c)
	return strings.Join(cColumns, "=?") + "=?", cValues, err
}

func streamCreateQuery(s *Stream) (string, []interface{}, error) {
	if s.Name == nil {
		return "", nil, ErrInvalidName
	}
	if s.Owner == nil {
		return "", nil, ErrInvalidQuery
	}
	sColumns, sValues, err := extractStream(s)
	if err != nil {
		return "", nil, err
	}

	// We create an ID for the connection. Guaranteed to be last element
	sColumns = append(sColumns, "id")
	sid := uuid.New().String()
	sValues = append(sValues, sid)
	s.ID = sid

	return strings.Join(sColumns, ","), sValues, err
}

func streamUpdateQuery(s *Stream) (string, []interface{}, error) {
	sColumns, sValues, err := extractStream(s)
	return strings.Join(sColumns, "=?") + "=?", sValues, err
}
