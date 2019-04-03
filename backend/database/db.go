package database

import (
	"errors"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// Details is used in groups, users, connections and streams to hold info
type Details struct {
	// The ID is used as a handle for all modification, and as such is also present in users
	ID          string  `json:"id,omitempty" db:"id"`
	Name        *string `json:"name" db:"name"`
	FullName    *string `json:"fullname" db:"fullname"`
	Description *string `json:"description" db:"description"`
	Avatar      *string `json:"avatar"`
}

// User holds a user's data
type User struct {
	Details

	PublicAccess *int `json:"public_access,omitempty" db:"public_access"`
	UserAccess   *int `json:"user_access,omitempty" db:"user_access"`

	Password *string `json:"password,omitempty" db:"password"`
}

// Group holds a group's details
type Group struct {
	Details
	Owner *string `json:"owner" db:"owner"`

	PublicAccess *int `json:"public_access,omitempty" db:"public_access"`
	UserAccess   *int `json:"user_access,omitempty" db:"user_access"`
}

type Connection struct {
	Details
	Owner *string `json:"owner" db:"owner"`

	APIKey *string `json:"apikey,omitempty" db:"apikey"`

	Settings      *string `json:"settings,omitempty" db:"settings"`
	SettingSchema *string `json:"setting_schema,omitempty" db:"setting_schema"`
}

type Stream struct {
	Details

	User       *string `json:"user,omitempty" db:"user"`
	Connection *string `json:"connection,omitempty" db:"connection"`
	Schema     *string `json:"schema,omitempty" db:"schema"`
	External   *string `json:"external,omitempty" db:"external"`
	Actor      *bool   `json:"actor,omitempty" db:"actor"`
	Access     *int    `json:"access,omitempty" db:"access"`
}

// ReadUserOptions gives options for reading a user
type ReadUserOptions struct {
	Avatar bool `json:"avatar,omitempty" schema:"avatar"`
}

// ReadGroupOptions gives options for reading
type ReadGroupOptions struct {
	Avatar bool `json:"avatar,omitempty" schema:"avatar"`
}

// ReadConnectionOptions gives options for reading
type ReadConnectionOptions struct {
	Avatar bool `json:"avatar,omitempty" schema:"avatar"`
}

// ReadStreamOptions gives options for reading
type ReadStreamOptions struct {
	Avatar bool `json:"avatar,omitempty" schema:"avatar"`
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
	ReadUser(name string, o *ReadUserOptions) (*User, error)
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

// Gets all pointer elements of a struct, and wherever the pointer isn't nil, adds it to the array
func extractPointers(o interface{}) (columns []string, values []interface{}) {
	v := reflect.ValueOf(o)
	k := v.Kind()
	for k == reflect.Ptr {
		v = reflect.Indirect(v)
		k = v.Kind()
	}
	t := v.Type()

	columns = make([]string, 0)
	values = make([]interface{}, 0)

	tot := v.NumField()
	for i := 0; i < tot; i++ {
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			// Only if it is a ptr do we continue, since that's all that we care about
			values = append(values, fieldValue.Interface())
			columns = append(columns, t.Field(i).Tag.Get("db"))
		}
	}

	return
}

// -------------------------------------------------------------------------------------
// Setting up reading and writing sql queries from objects
// -------------------------------------------------------------------------------------

func extractDetails(d *Details) (columns []string, values []interface{}, err error) {
	if d.Name != nil {
		if err = ValidName(*d.Name); err != nil {
			return
		}
	}
	if d.Avatar != nil {
		if err = ValidAvatar(*d.Avatar); err != nil {
			return
		}
	}
	columns, values = extractPointers(d)

	return
}

func extractGroup(g *Group) (groupColumns []string, groupValues []interface{}, err error) {
	if g.Owner != nil {
		if err = ValidName(*g.Owner); err != nil {
			return
		}
	}

	if g.PublicAccess != nil {
		if err = ValidGroupAccessLevel(*g.PublicAccess); err != nil {
			return
		}
	}
	if g.UserAccess != nil {
		if err = ValidGroupAccessLevel(*g.UserAccess); err != nil {
			return
		}
	}

	groupColumns, groupValues, err = extractDetails(&g.Details)
	if err != nil {
		return nil, nil, err
	}
	c2, g2 := extractPointers(g)
	groupColumns = append(groupColumns, c2...)
	groupValues = append(groupValues, g2...)

	return
}

func extractUser(u *User) (userColumns []string, userValues []interface{}, err error) {
	userColumns, userValues, err = extractDetails(&u.Details)
	if err != nil {
		return
	}
	if u.PublicAccess != nil {
		if err = ValidGroupAccessLevel(*u.PublicAccess); err != nil {
			return
		}
	}
	if u.UserAccess != nil {
		if err = ValidGroupAccessLevel(*u.UserAccess); err != nil {
			return
		}
	}
	if u.Password != nil {
		var password string
		password, err = HashPassword(*u.Password)
		if err != nil {
			return
		}
		u.Password = &password
	}
	c2, g2 := extractPointers(u)
	userColumns = append(userColumns, c2...)
	userValues = append(userValues, g2...)

	if len(userColumns) < 1 {
		err = ErrNoUpdate
	}
	return
}

func extractConnection(c *Connection) (cColumns []string, cValues []interface{}, err error) {
	cColumns, cValues, err = extractDetails(&c.Details)
	if err != nil {
		return
	}
	if c.Owner != nil {
		if err = ValidName(*c.Owner); err != nil {
			return
		}
	}

	if c.APIKey != nil {

		if *c.APIKey != "" {
			// Anything else we replace with a new API key
			var apikey string
			apikey, err = GenerateKey(15)
			if err != nil {
				return
			}
			c.APIKey = &apikey // Write the API key back to the connection object
		}

	}

	c2, g2 := extractPointers(c)
	cColumns = append(cColumns, c2...)
	cValues = append(cValues, g2...)

	return
}

func extractStream(s *Stream) (sColumns []string, sValues []interface{}, err error) {
	sColumns, sValues, err = extractDetails(&s.Details)
	if err != nil {
		return
	}
	if s.User != nil {
		if err = ValidName(*s.User); err != nil {
			return
		}
	}
	c2, g2 := extractPointers(s)
	sColumns = append(sColumns, c2...)
	sValues = append(sValues, g2...)

	return
}

// Insert the right amount of question marks for the given query
func qQ(size int) string {
	s := strings.Repeat("?,", size)
	return s[:len(s)-1]
}

func userCreateQuery(u *User) (string, []interface{}, error) {
	if u.Name == nil {
		return "", nil, ErrInvalidName
	}
	if u.Password == nil || "" == *u.Password {
		return "", nil, ErrNoPasswordGiven
	}

	userColumns, userValues, err := extractUser(u)
	if err != nil {
		return "", nil, err
	}

	return strings.Join(userColumns, ","), userValues, err
}

func userUpdateQuery(u *User) (string, []interface{}, error) {
	if err := ValidName(u.ID); err != nil {
		return "", nil, err
	}
	uColumns, userValues, err := extractUser(u)
	if err != nil {
		return "", nil, err
	}

	userColumns := strings.Join(uColumns, "=?,") + "=?"

	userValues = append(userValues, u.ID)

	return userColumns, userValues, err
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
		// We want the API key to always be set on create
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
	if s.User == nil && s.Connection == nil {
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
