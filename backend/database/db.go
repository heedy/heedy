package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/heedy/heedy/backend/assets"
)

type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	t := fmt.Sprintf("\"%s\"", d.String())
	return []byte(t), nil
}

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

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
	total = strings.TrimSpace(total)
	if len(total) == 0 {
		s.Scopes = []string{}
	} else {
		s.Scopes = strings.Split(total, " ")
	}
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

func appParentScope(s string) string {
	r := strings.SplitN(s, ":", 2)
	return r[0]
}

// AppScopeArray works with app scopes, which have different details than object scopes
type AppScopeArray struct {
	ScopeArray
}

// Update cleans out the scopes to remove repeated items
func (s *AppScopeArray) Update() {
	scopeMap := make(map[string]bool)
	for _, v := range s.Scopes {
		scopeMap[v] = true
	}

	// Now for each scope, check if there is a wildcard, meaning that
	// self.objects encompasses self.objects:read

	s.scopeMap = make(map[string]bool)
	if _, ok := scopeMap["*"]; ok {
		s.scopeMap["*"] = true
		s.Scopes = []string{"*"}
		return
	}

	for _, v := range s.Scopes {
		if _, ok := scopeMap[appParentScope(v)]; !ok {
			s.scopeMap[v] = true
		}
	}

	s.Scopes = make([]string, 0, len(s.scopeMap))
	for k := range s.scopeMap {
		s.Scopes = append(s.Scopes, k)
	}
}

// HasScope checks if the given scope is present
func (s *AppScopeArray) HasScope(sv string) (ok bool) {
	if s.scopeMap == nil {
		s.Update()
	}
	_, ok = s.scopeMap[sv]
	if !ok {
		_, ok = s.scopeMap[appParentScope(sv)]
		if !ok {
			_, ok = s.scopeMap["*"]
		}
	}
	return
}

// JSONObject represents a json column in the table. To handle it correctly, we need to manually scan it
// and output the relevant values
type JSONObject map[string]interface{}

func (s *JSONObject) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &s)
		return nil
	case string:
		json.Unmarshal([]byte(v), &s)
		return nil
	default:
		return fmt.Errorf("Can't unmarshal json object, unsupported type: %T", v)
	}
}
func (s *JSONObject) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Details is used in groups, users, apps and objects to hold info
type Details struct {
	// The ID is used as a handle for all modification, and as such is also present in users
	ID          string  `json:"id,omitempty" db:"id"`
	Name        *string `json:"name,omitempty" db:"name"`
	Description *string `json:"description,omitempty" db:"description"`
	Icon        *string `json:"icon,omitempty" db:"icon"`
}

// User holds a user's data
type User struct {
	Details

	UserName *string `json:"username" db:"username"`

	PublicRead *bool `json:"public_read" db:"public_read"`
	UsersRead  *bool `json:"users_read" db:"users_read"`

	Password *string `json:"password,omitempty" db:"password"`
}

type App struct {
	Details
	Owner  *string `json:"owner" db:"owner"`
	Plugin *string `json:"plugin" db:"plugin"`
	Type   *string `json:"type" db:"type"`

	Enabled *bool `json:"enabled,omitempty" db:"enabled"`

	AccessToken    *string `json:"access_token,omitempty" db:"access_token"`
	CreatedDate    Date    `json:"created_date,omitempty" db:"created_date"`
	LastAccessDate *Date   `json:"last_access_date" db:"last_access_date"`

	Scopes *AppScopeArray `json:"scopes" db:"scopes"`

	Settings       *JSONObject `json:"settings" db:"settings"`
	SettingsSchema *JSONObject `json:"settings_schema" db:"settings_schema"`
}

type Object struct {
	Details

	Owner *string `json:"owner,omitempty" db:"owner"`
	App   *string `json:"app" db:"app"`

	Key *string `json:"key" db:"key"`

	Type *string     `json:"type,omitempty" db:"type"`
	Meta *JSONObject `json:"meta,omitempty" db:"meta"`

	CreatedDate  *Date `json:"created_date,omitempty" db:"created_date"`
	LastModified *Date `json:"last_modified" db:"last_modified"`

	Scopes *ScopeArray `json:"scopes" db:"scopes"`

	// The access array, giving the permissions the cuurently logged in thing has
	// It is generated manually for each read query, it does not exist in the database.
	Access ScopeArray `json:"access,omitempty" db:"access"`
}

func (s *Object) String() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// ReadUserOptions gives options for reading a user
type ReadUserOptions struct {
	Icon bool `json:"icon,omitempty" schema:"icon"`
}

// ReadAppOptions gives options for reading
type ReadAppOptions struct {
	Icon        bool `json:"icon,omitempty" schema:"icon"`
	AccessToken bool `json:"token,omitempty" schema:"token"` // using "token" instead of access_token, since the API uses access_token param
}

// ReadObjectOptions gives options for reading
type ReadObjectOptions struct {
	Icon bool `json:"icon,omitempty" schema:"icon"`
}

type ListUsersOptions struct {
}

// ListObjectsOptions shows the options for listing objects
type ListObjectsOptions struct {
	// Whether to include icons
	Icon *bool `json:"icon,omitempty" schema:"icon"`
	// Limit results to the given user's objects.
	UserName *string `json:"username,omitempty" schema:"username"`
	// Limit the results to the given app's objects
	App *string `json:"app,omitempty" schema:"app"`
	// Get objects with the given key
	Key *string `json:"key,omitempty" schema:"key"`
	// Limit results to objects of the given type
	Type *string `json:"type,omitempty" schema:"type"`
	// Maximum number of results to return
	Limit *int `json:"limit,omitempty" schema:"limit"`

	// Whether to include shared objects (not belonging to the user)
	// This is only allowed for user==current user
	Shared *bool
}

// ListAppOptions holds the options associated with listing apps
type ListAppOptions struct {
	// Whether to include icons
	Icon *bool `json:"icon,omitempty" schema:"icon"`
	// Limit results to the given user's apps
	User *string `json:"user,omitempty" schema:"user"`
	// Find the apps with the given plugin key
	Plugin *string `json:"plugin,omitempty" schema:"plugin"`
}

type DBType int

const (
	PublicType DBType = iota
	AppType
	UserType
	AdminType
)

// DB represents the database. This interface is implemented in many ways:
//	once for admin
//	once for users
//	once for apps
//	once for public
type DB interface {
	AdminDB() *AdminDB // Returns the underlying administrative database
	ID() string        // This is an identifier for the database
	Type() DBType      // Returns the database type

	// Currently logged in user
	// User() (*User, error)

	CreateUser(u *User) error
	ReadUser(name string, o *ReadUserOptions) (*User, error)
	UpdateUser(u *User) error
	DelUser(name string) error
	ListUsers(o *ListUsersOptions) ([]*User, error)

	CreateApp(c *App) (string, string, error)
	ReadApp(cid string, o *ReadAppOptions) (*App, error)
	UpdateApp(c *App) error
	DelApp(cid string) error
	ListApps(o *ListAppOptions) ([]*App, error)

	CanCreateObject(s *Object) error
	CreateObject(s *Object) (string, error)
	ReadObject(id string, o *ReadObjectOptions) (*Object, error)
	UpdateObject(s *Object) error
	DelObject(id string) error

	ShareObject(objectid, userid string, sa *ScopeArray) error
	UnshareObjectFromUser(objectid, userid string) error
	UnshareObject(objectid string) error
	GetObjectShares(objectid string) (m map[string]*ScopeArray, err error)

	ListObjects(o *ListObjectsOptions) ([]*Object, error)
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
	ErrNotFound        = errors.New("not_found: The selected reobject was not found")
	ErrNoUpdate        = errors.New("Nothing to update")
	ErrNoPasswordGiven = errors.New("A user cannot have an empty password")
	ErrUserNotFound    = errors.New("User was not found")
	ErrInvalidUserName = errors.New("Invalid Username")
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

	if d.Icon != nil {
		if err = ValidIcon(*d.Icon); err != nil {
			return
		}
	}

	columns, values = extractPointers(d)

	return
}

func extractUser(u *User) (userColumns []string, userValues []interface{}, err error) {
	userColumns, userValues, err = extractDetails(&u.Details)
	if err != nil {
		return
	}
	if u.UserName != nil {
		if err = ValidUserName(*u.UserName); err != nil {
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

func extractApp(c *App) (cColumns []string, cValues []interface{}, err error) {
	// We don't allow modifying last access date
	c.LastAccessDate = nil
	cColumns, cValues, err = extractDetails(&c.Details)
	if err != nil {
		return
	}
	if c.Owner != nil {
		if err = ValidUserName(*c.Owner); err != nil {
			return
		}
	}

	noToken := false
	if c.AccessToken != nil {

		if *c.AccessToken != "" {
			// Anything else we replace with a new token
			var token string
			token, err = GenerateKey(15)
			if err != nil {
				return
			}
			c.AccessToken = &token // Write the token back to the app object
		} else {
			noToken = true
			// Make the pointer not extact
			c.AccessToken = nil
			// set the access token to NULL
			cColumns = append(cColumns, "access_token")
			cValues = append(cValues, nil)
		}

	}

	c2, g2 := extractPointers(c)
	cColumns = append(cColumns, c2...)
	cValues = append(cValues, g2...)

	if noToken {
		// Needed to stop generating a key for apps that don't want one
		emptystring := ""
		c.AccessToken = &emptystring
	}

	return
}

func extractObject(s *Object) (sColumns []string, sValues []interface{}, err error) {
	sColumns, sValues, err = extractDetails(&s.Details)
	if err != nil {
		return
	}
	if s.Owner != nil {
		if err = ValidUserName(*s.Owner); err != nil {
			return
		}
	}
	if len(s.Access.Scopes) > 0 {
		err = ErrBadQuery("The access field is auto-generated from permissions - it cannot be set directly")
	}

	c2, g2 := extractPointers(s)
	sColumns = append(sColumns, c2...)
	sValues = append(sValues, g2...)

	return
}

// Insert the right amount of question marks for the given query
func QQ(size int) string {
	s := strings.Repeat("?,", size)
	return s[:len(s)-1]
}

func sqlIn(s string, v []string) string {
	return fmt.Sprintf(s, "'"+strings.Join(v, "', '")+"'")
}

func userCreateQuery(u *User) (string, []interface{}, error) {
	if u.UserName == nil {
		return "", nil, ErrInvalidUserName
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
	if err := ValidUserName(u.ID); err != nil {
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

func appCreateQuery(c *App) (string, []interface{}, error) {
	if c.Name == nil {
		return "", nil, ErrInvalidName
	}
	if c.Owner == nil {
		return "", nil, ErrInvalidQuery
	}
	cColumns, cValues, err := extractApp(c)
	if err != nil {
		return "", nil, err
	}

	if c.AccessToken == nil {
		accessToken, err := GenerateKey(15)
		if err != nil {
			return "", nil, err
		}
		c.AccessToken = &accessToken

		// Add the token to things we set
		cColumns = append(cColumns, "access_token")
		cValues = append(cValues, accessToken)
	}

	// We create an ID for the app. Guaranteed to be last element
	cColumns = append(cColumns, "id")
	cid := uuid.New().String()
	cValues = append(cValues, cid)
	c.ID = cid

	return strings.Join(cColumns, ","), cValues, err
}

func appUpdateQuery(c *App) (string, []interface{}, error) {
	cColumns, cValues, err := extractApp(c)
	if len(cValues) == 0 {
		return "", nil, ErrNoUpdate
	}
	return strings.Join(cColumns, "=?,") + "=?", cValues, err
}

func objectCreateQuery(c *assets.Configuration, s *Object) (string, []interface{}, error) {
	var err error
	if s.Name == nil {
		return "", nil, ErrInvalidName
	}
	if s.Owner == nil && s.App == nil {
		return "", nil, ErrBadQuery("You must specify either an owner or a app to which the object should belong")
	}
	if s.App != nil && s.Owner != nil {
		return "", nil, ErrBadQuery("When creating a object for a app, you must not specify an owner")
	}
	if s.Type == nil {
		return "", nil, ErrBadQuery("Must specify a object type")
	}
	if s.Meta != nil {
		err = c.ValidateObjectMetaWithDefaults(*s.Type, *s.Meta)
	} else {
		// Validate will set up default meta values
		m := JSONObject{}
		err = c.ValidateObjectMetaWithDefaults(*s.Type, m)
		s.Meta = &m
	}
	if err != nil {
		return "", nil, err
	}

	sColumns, sValues, err := extractObject(s)
	if err != nil {
		return "", nil, err
	}

	// We create an ID for the app. Guaranteed to be last element
	sColumns = append(sColumns, "id")
	sid := uuid.New().String()
	sValues = append(sValues, sid)
	s.ID = sid

	return strings.Join(sColumns, ","), sValues, err
}

// The object s is assumed to have the underlying object type added in.
func objectUpdateQuery(c *assets.Configuration, s *Object, objectType string) (string, []interface{}, error) {
	sColumns, sValues, err := extractObject(s)
	if s.Type != nil {
		return "", nil, ErrBadQuery("Modifying a object type is not supported")
	}
	if len(sValues) == 0 {
		return "", nil, ErrNoUpdate
	}
	if s.Meta != nil && err != nil {
		err = c.ValidateObjectMeta(*s.Type, (*map[string]interface{})(s.Meta))
	}
	return strings.Join(sColumns, "=?,") + "=?", sValues, err
}

func listObjectsQuery(o *ListObjectsOptions) (string, []interface{}, error) {
	sColumns := make([]string, 0)
	sValues := make([]interface{}, 0)
	pretext := ""
	if o != nil {

		if o.UserName != nil {
			sColumns = append(sColumns, "owner")
			sValues = append(sValues, *o.UserName)
		}
		if o.App != nil {
			if *o.App == "none" {
				pretext = "app IS NULL"
			} else {
				sColumns = append(sColumns, "app")
				sValues = append(sValues, *o.App)
			}
		}
		if o.Type != nil {
			sColumns = append(sColumns, "type")
			sValues = append(sValues, *o.Type)
		}
		if o.Key != nil {
			sColumns = append(sColumns, "key")
			sValues = append(sValues, *o.Key)
		}
	}
	if len(sColumns) == 0 {
		if len(pretext) == 0 {
			return "1=1", sValues, nil
		}
		return pretext, sValues, nil
	}
	if len(pretext) == 0 {
		return strings.Join(sColumns, "=? AND ") + "=?", sValues, nil
	}

	return pretext + " AND " + strings.Join(sColumns, "=? AND ") + "=?", sValues, nil
}
