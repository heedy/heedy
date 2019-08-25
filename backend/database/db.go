package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/heedy/heedy/backend/assets"
)

// ScopeArray represents a json column in a table. To handle it correctly, we need to manually scan it
// and output a value.
type ScopeArray struct {
	Scopes   []string
	scopeMap map[string]bool
}

// Update cleans out the scopes to remove repeated items
func (s *ScopeArray) Update() {
	s.scopeMap = make(map[string]bool)
	for _, v := range s.Scopes {
		s.scopeMap[v] = true
	}

	if _, ok := s.scopeMap["*"]; ok {
		s.Scopes = []string{"*"}
		return
	}
	s.Scopes = make([]string, 0, len(s.scopeMap))
	for k := range s.scopeMap {
		s.Scopes = append(s.Scopes, k)
	}
}

func (s *ScopeArray) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &s.Scopes)
		s.Update()
		return nil
	case string:
		json.Unmarshal([]byte(v), &s.Scopes)
		s.Update()
		return nil
	default:
		return fmt.Errorf("Can't scan scope array, unsupported type: %T", v)
	}
}

func (s *ScopeArray) String() string {
	return strings.Join(s.Scopes, " ")
}

func (s *ScopeArray) MarshalJSON() ([]byte, error) {

	return json.Marshal(s.String())
}

func (s *ScopeArray) UnmarshalJSON(b []byte) error {
	var total string
	err := json.Unmarshal(b, &total)
	s.Scopes = strings.Split(total, " ")
	s.Update()
	return err
}

func (s *ScopeArray) Value() (driver.Value, error) {
	return json.Marshal(s.Scopes)
}

// HasScope checks if the given scope is present
func (s *ScopeArray) HasScope(sv string) (ok bool) {
	if s.scopeMap == nil {
		s.Update()
	}
	_, ok = s.scopeMap[sv]
	if !ok {
		_, ok = s.scopeMap["*"]
	}
	return
}

// SourceMeta represents a json column in the table. To handle it correctly, we need to manually scan it
// and output the relevant values
type SourceMeta map[string]interface{}

func (s *SourceMeta) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &s)
		return nil
	case string:
		json.Unmarshal([]byte(v), &s)
		return nil
	default:
		return fmt.Errorf("Can't unmarshal source meta, unsupported type: %T", v)
	}
}
func (s *SourceMeta) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Details is used in groups, users, connections and sources to hold info
type Details struct {
	// The ID is used as a handle for all modification, and as such is also present in users
	ID          string  `json:"id,omitempty" db:"id"`
	Name        *string `json:"name" db:"name"`
	FullName    *string `json:"fullname" db:"fullname"`
	Description *string `json:"description" db:"description"`
	Avatar      *string `json:"avatar" db:"avatar"`

	// The access array, giving the permissions the cuurently logged in thing has
	// It is generated manually for each read query, it does not exist in the database.
	Access ScopeArray `json:"access,omitempty" db:"access"`
}

// User holds a user's data
type User struct {
	Details

	PublicRead *bool `json:"public_read" db:"public_read"`
	UsersRead  *bool `json:"users_read" db:"users_read"`

	Password *string `json:"password,omitempty" db:"password"`
}

type Connection struct {
	Details
	Owner *string `json:"owner" db:"owner"`

	APIKey *string `json:"apikey,omitempty" db:"apikey"`

	Scopes *ScopeArray `json:"scopes" db:"scopes"`

	Settings      *string `json:"settings" db:"settings"`
	SettingSchema *string `json:"setting_schema" db:"setting_schema"`
}

type Source struct {
	Details

	Owner      *string `json:"owner,omitempty" db:"owner"`
	Connection *string `json:"connection,omitempty" db:"connection"`

	Type *string     `json:"type,omitempty" db:"type"`
	Meta *SourceMeta `json:"meta,omitempty" db:"meta"`

	Scopes *ScopeArray `json:"scopes" db:"scopes"`
}

func (s *Source) String() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// ReadUserOptions gives options for reading a user
type ReadUserOptions struct {
	Avatar bool `json:"avatar,omitempty" schema:"avatar"`
}

// ReadConnectionOptions gives options for reading
type ReadConnectionOptions struct {
	Avatar bool `json:"avatar,omitempty" schema:"avatar"`
	APIKey bool `json:"apikey,omitempty" schema:"apikey"`
}

// ReadSourceOptions gives options for reading
type ReadSourceOptions struct {
	Avatar bool `json:"avatar,omitempty" schema:"avatar"`
}

// ListSourcesOptions shows the options for listing sources
type ListSourcesOptions struct {
	// Whether to include avatars
	Avatar *bool `json:"avatar,omitempty" schema:"avatar"`
	// Limit results to the given user's sources.
	User *string `json:"user,omitempty" schema:"user"`
	// Limit the results to the given connection's sources
	Connection *string `json:"connection,omitempty" schema:"connection"`
	// Limit results to sources of the given type
	Type *string `json:"type,omitempty" schema:"type"`
	// Maximum number of results to return
	Limit *int `json:"limit,omitempty" schema:"limit"`

	// Whether to include shared sources (not belonging to the user)
	// This is only allowed for user==current user
	Shared *bool
}

// ListConnectionOptions holds the options associated with listing connections
type ListConnectionOptions struct {
	// Whether to include avatars
	Avatar *bool `json:"avatar,omitempty" schema:"avatar"`
	// Limit results to the given user's connections
	User *string `json:"user,omitempty" schema:"user"`
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
	// User() (*User, error)

	CreateUser(u *User) error
	ReadUser(name string, o *ReadUserOptions) (*User, error)
	UpdateUser(u *User) error
	DelUser(name string) error

	CreateConnection(c *Connection) (string,string,error)
	ReadConnection(cid string, o *ReadConnectionOptions) (*Connection,error)
	UpdateConnection(c *Connection) error
	DelConnection(cid string) error
	ListConnections(o *ListConnectionOptions) ([]*Connection,error)

	CanCreateSource(s *Source) error
	CreateSource(s *Source) (string, error)
	ReadSource(id string, o *ReadSourceOptions) (*Source, error)
	UpdateSource(s *Source) error
	DelSource(id string) error

	ShareSource(sourceid, userid string, sa *ScopeArray) error
	UnshareSourceFromUser(sourceid, userid string) error
	UnshareSource(sourceid string) error
	GetSourceShares(sourceid string) (m map[string]*ScopeArray, err error)

	ListSources(o *ListSourcesOptions) ([]*Source,error)
}

func ErrAccessDenied(err string, args ...interface{}) error {
	s := fmt.Sprintf(err, args...)
	return fmt.Errorf("access_denied: %s", s)
}
func ErrBadQuery(err string, args ...interface{}) error {
	s := fmt.Sprintf(err, args...)
	return fmt.Errorf("bad_query: %s", s)
}

var (
	ErrNotFound        = errors.New("not_found: The selected resource was not found")
	ErrNoUpdate        = errors.New("Nothing to update")
	ErrNoPasswordGiven = errors.New("A user cannot have an empty password")
	ErrUserNotFound    = errors.New("User was not found")
	ErrInvalidName     = errors.New("Invalid name")
	ErrInvalidQuery    = errors.New("Invalid query")
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
	if len(d.Access.Scopes) > 0 {
		err = ErrBadQuery("The access field is auto-generated from permissions - it cannot be set directly")
	}
	columns, values = extractPointers(d)

	return
}

func extractUser(u *User) (userColumns []string, userValues []interface{}, err error) {
	userColumns, userValues, err = extractDetails(&u.Details)
	if err != nil {
		return
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

func extractSource(s *Source) (sColumns []string, sValues []interface{}, err error) {
	sColumns, sValues, err = extractDetails(&s.Details)
	if err != nil {
		return
	}
	if s.Owner != nil {
		if err = ValidName(*s.Owner); err != nil {
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

func sqlIn(s string, v []string) string {
	return fmt.Sprintf(s, "'"+strings.Join(v, "', '")+"'")
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
	if len(uColumns) == 0 {
		return "", nil, ErrNoUpdate
	}

	userColumns := strings.Join(uColumns, "=?,") + "=?"

	userValues = append(userValues, u.ID)

	return userColumns, userValues, err
}

func connectionCreateQuery(c *Connection) (string, []interface{}, error) {
	if c.Name == nil {
		return "", nil, ErrInvalidName
	}
	if c.Owner == nil {
		return "", nil, ErrInvalidQuery
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
	if len(cValues) == 0 {
		return "", nil, ErrNoUpdate
	}
	return strings.Join(cColumns, "=?,") + "=?", cValues, err
}

func sourceCreateQuery(c *assets.Configuration, s *Source) (string, []interface{}, error) {
	var err error
	if s.Name == nil {
		return "", nil, ErrInvalidName
	}
	if s.Owner == nil && s.Connection == nil {
		return "", nil, ErrBadQuery("You must specify either an owner or a connection to which the source should belong")
	}
	if s.Connection != nil && s.Owner != nil {
		return "", nil, ErrBadQuery("When creating a source for a connection, you must not specify an owner")
	}
	if s.Type == nil {
		return "", nil, ErrBadQuery("Must specify a source type")
	}
	if s.Meta != nil {
		err = c.ValidateSourceMetaWithDefaults(*s.Type, *s.Meta)
	} else {
		m := map[string]interface{}{}
		err = c.ValidateSourceMetaWithDefaults(*s.Type, m)
	}
	if err != nil {
		return "", nil, err
	}

	sColumns, sValues, err := extractSource(s)
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

// The source s is assumed to have the underlying source type added in.
func sourceUpdateQuery(c *assets.Configuration, s *Source, sourceType string) (string, []interface{}, error) {
	sColumns, sValues, err := extractSource(s)
	if s.Type != nil {
		return "", nil, ErrBadQuery("Modifying a source type is not supported")
	}
	if len(sValues) == 0 {
		return "", nil, ErrNoUpdate
	}
	if s.Meta != nil && err != nil {
		err = c.ValidateSourceMeta(*s.Type, (*map[string]interface{})(s.Meta))
	}
	return strings.Join(sColumns, "=?,") + "=?", sValues, err
}

func listSourcesQuery(o *ListSourcesOptions) (string,[]interface{},error) {
	sColumns := make([]string,0)
	sValues := make([]interface{},0)
	pretext := ""
	if o!=nil {

		if o.User!=nil {
			sColumns = append(sColumns,"owner")
			sValues = append(sValues,*o.User)
		}
		if o.Connection!=nil {
			if *o.Connection == "none" {
				pretext = "connection IS NULL"
			} else {
				sColumns = append(sColumns,"connection")
				sValues = append(sValues,*o.Connection)
			}
		}
		if o.Type!=nil {
			sColumns = append(sColumns,"type")
			sValues = append(sValues,*o.Type)
		}
	}
	if len(sColumns)==0 {
		if len(pretext)==0 {
			return "1=1",sValues,nil
		}
		return pretext,sValues,nil
	}
	if len(pretext)==0 {
		return strings.Join(sColumns, "=? AND ") + "=?", sValues, nil
	}

	return pretext + " AND " + strings.Join(sColumns, "=? AND ") + "=?", sValues, nil
}