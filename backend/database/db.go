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
	"github.com/heedy/heedy/backend/database/dbutil"
)

// ScopeArray represents a json column in a table. To handle it correctly, we need to manually scan it
// and output a value.
type ScopeArray struct {
	Scope    []string
	scopeMap map[string]bool
}

// Update cleans out the scope to remove repeated items
func (s *ScopeArray) Update() {
	s.scopeMap = make(map[string]bool)
	for _, v := range s.Scope {
		s.scopeMap[v] = true
	}

	if _, ok := s.scopeMap["*"]; ok {
		s.Scope = []string{"*"}
		return
	}
	s.Scope = make([]string, 0, len(s.scopeMap))
	for k := range s.scopeMap {
		s.Scope = append(s.Scope, k)
	}
}

func (s *ScopeArray) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &s.Scope)
		s.Update()
		return nil
	case string:
		json.Unmarshal([]byte(v), &s.Scope)
		s.Update()
		return nil
	default:
		return fmt.Errorf("Can't scan scope array, unsupported type: %T", v)
	}
}

func (s *ScopeArray) String() string {
	return strings.Join(s.Scope, " ")
}

func (s *ScopeArray) MarshalJSON() ([]byte, error) {

	return json.Marshal(s.String())
}

func (s *ScopeArray) Load(total string) {
	s.Scope = strings.Fields(total)
	s.Update()
}

func (s *ScopeArray) UnmarshalJSON(b []byte) error {
	var total string
	err := json.Unmarshal(b, &total)
	s.Load(total)
	return err
}

func (s *ScopeArray) Value() (driver.Value, error) {
	return json.Marshal(s.Scope)
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

// AppScopeArray works with app scope, which have different details than object scope
type AppScopeArray struct {
	ScopeArray
}

// Update cleans out the scope to remove repeated items
func (s *AppScopeArray) Update() {
	scopeMap := make(map[string]bool)
	for _, v := range s.Scope {
		scopeMap[v] = true
	}

	// Now for each scope, check if there is a wildcard, meaning that
	// self.objects encompasses self.objects:read

	s.scopeMap = make(map[string]bool)
	if _, ok := scopeMap["*"]; ok {
		s.scopeMap["*"] = true
		s.Scope = []string{"*"}
		return
	}

	for _, v := range s.Scope {
		if _, ok := scopeMap[appParentScope(v)]; !ok {
			s.scopeMap[v] = true
		}
	}

	s.Scope = make([]string, 0, len(s.scopeMap))
	for k := range s.scopeMap {
		s.Scope = append(s.Scope, k)
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
	Plugin *string `json:"plugin,omitempty" db:"plugin"`

	Enabled *bool `json:"enabled,omitempty" db:"enabled"`

	AccessToken    *string      `json:"access_token,omitempty" db:"access_token"`
	CreatedDate    dbutil.Date  `json:"created_date,omitempty" db:"created_date"`
	LastAccessDate *dbutil.Date `json:"last_access_date" db:"last_access_date"`

	Scope *AppScopeArray `json:"scope" db:"scope"`

	Settings       *dbutil.JSONObject `json:"settings" db:"settings"`
	SettingsSchema *dbutil.JSONObject `json:"settings_schema" db:"settings_schema"`
}

type Object struct {
	Details

	Owner *string `json:"owner,omitempty" db:"owner"`
	App   *string `json:"app" db:"app"`

	Tags *dbutil.StringArray `json:"tags,omitempty" db:"tags"`

	Key *string `json:"key,omitempty" db:"key"`

	Type *string            `json:"type,omitempty" db:"type"`
	Meta *dbutil.JSONObject `json:"meta,omitempty" db:"meta"`

	CreatedDate  *dbutil.Date `json:"created_date,omitempty" db:"created_date"`
	ModifiedDate *dbutil.Date `json:"modified_date" db:"modified_date"`

	// The scope the owner has to the object. This allows apps to control objects belonging to them.
	OwnerScope *ScopeArray `json:"owner_scope,omitempty" db:"owner_scope"`

	// The access array, giving the permissions the currently logged in thing has
	// It is generated manually for each read query, it does not exist in the database.
	Access ScopeArray `json:"access,omitempty" db:"access"`
}

func (s *Object) String() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

type UserSession struct {
	SessionID      string      `db:"sessionid" json:"sessionid"`
	Description    string      `db:"description" json:"description"`
	LastAccessDate dbutil.Date `db:"last_access_date" json:"last_access_date"`
	CreatedDate    dbutil.Date `db:"created_date" json:"created_date"`
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
	ReadUserOptions
}

// ListObjectsOptions shows the options for listing objects
type ListObjectsOptions struct {
	ReadObjectOptions

	// Limit results to the given user's objects.
	Owner *string `json:"owner,omitempty" schema:"owner"`
	// Limit the results to the given app's objects
	App *string `json:"app,omitempty" schema:"app"`
	// Get by plugin key
	Key *string `json:"key,omitempty" schema:"key"`
	// Get objects with the given tags
	Tags *string `json:"tags,omitempty" schema:"tags"`
	// Limit results to objects of the given type
	Type *string `json:"type,omitempty" schema:"type"`
	// Maximum number of results to return
	Limit *int `json:"limit,omitempty" schema:"limit"`

	// Whether to include shared objects (not belonging to the user)
	// This is only allowed for user==current user
	Shared bool
}

// ListAppOptions holds the options associated with listing apps
type ListAppOptions struct {
	ReadAppOptions

	// Limit results to the given user's apps
	Owner *string `json:"owner,omitempty" schema:"owner"`
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

	ListUserSessions(name string) ([]UserSession, error)
	DelUserSession(name, id string) error

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

	ReadUserSettings(username string) (map[string]map[string]interface{}, error)
	UpdateUserPluginSettings(username string, plugin string, preferences map[string]interface{}) error
	ReadUserPluginSettings(username string, plugin string) (map[string]interface{}, error)
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
	ErrNoUpdate        = errors.New("nop: Nothing to update")
	ErrNoPasswordGiven = errors.New("bad_request: A user cannot have an empty password")
	ErrUserNotFound    = errors.New("not_found: User was not found")
	ErrInvalidUserName = errors.New("bad_request: Invalid Username")
	ErrInvalidName     = errors.New("bad_request: Invalid name")
	ErrInvalidQuery    = errors.New("invalid_query: Invalid query")
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
	if len(s.Access.Scope) > 0 {
		err = ErrBadQuery("The access field is auto-generated from permissions - it cannot be set directly")
	}
	skv := s.Key
	if skv != nil && *skv == "" {
		sColumns = append(sColumns, "key")
		sValues = append(sValues, nil)
		s.Key = nil
	}

	c2, g2 := extractPointers(s)
	sColumns = append(sColumns, c2...)
	sValues = append(sValues, g2...)

	s.Key = skv

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
		return "", nil, ErrBadQuery("An app must have an owner")
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
	if s.Key != nil && *s.Key == "" {
		return "", nil, ErrBadQuery("Object key can't be empty string")
	}
	if s.Meta != nil {
		err = c.ValidateObjectMetaWithDefaults(*s.Type, *s.Meta)
	} else {
		// Validate will set up default meta values
		m := dbutil.JSONObject{}
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
	metav := s.Meta
	if metav != nil {
		err := c.ValidateObjectMetaUpdate(objectType, *metav)
		if err != nil {
			return "", nil, err
		}
		// Stop the meta from being extracted from the object, since the update query
		// needs to handle it manually
		s.Meta = nil
	}
	sColumns, sValues, err := extractObject(s)
	if s.Type != nil {
		return "", nil, ErrBadQuery("Modifying a object type is not supported")
	}

	if metav == nil {
		if len(sValues) == 0 {
			return "", nil, ErrNoUpdate
		}
		return strings.Join(sColumns, "=?,") + "=?", sValues, err
	}

	// Handle meta values
	deletes := make([]interface{}, 0)
	adds := make([]interface{}, 0)
	for k, v := range *metav {
		if v == nil {
			deletes = append(deletes, "$."+k)
		} else {
			jsonvalue, err := json.Marshal(v)
			if err != nil {
				return "", nil, err
			}
			adds = append(adds, "$."+k, string(jsonvalue))
		}
	}
	if len(deletes) == 0 && len(adds) == 0 {
		if len(sValues) == 0 {
			return "", nil, ErrNoUpdate
		}
		return strings.Join(sColumns, "=?,") + "=?", sValues, err
	}

	metaq := "json(meta)"
	if len(deletes) > 0 {
		sValues = append(sValues, deletes...)
		metaq = fmt.Sprintf("json_remove(%s,%s)", metaq, QQ(len(deletes)))
	}
	if len(adds) > 0 {
		sValues = append(sValues, adds...)
		metaq = "json_set(" + metaq
		// Add the right number of qqs
		for i := 0; i < len(adds)/2; i++ {
			metaq += ",?,json(?)"
		}
		metaq += ")"
	}
	if len(sColumns) == 0 {
		return "meta=" + metaq, sValues, err
	}

	return strings.Join(sColumns, "=?,") + "=?, meta=" + metaq, sValues, err
}

func listObjectsQuery(o *ListObjectsOptions) (string, []interface{}, error) {
	sColumns := make([]string, 0)
	sValues := make([]interface{}, 0)
	pretext := ""
	if o != nil {

		if o.Owner != nil {
			sColumns = append(sColumns, "owner")
			sValues = append(sValues, *o.Owner)
		}
		if o.App != nil {
			if *o.App == "" {
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
			if *o.Key == "" {
				if pretext != "" {
					pretext += " AND "
				}
				pretext += "key IS NULL"
			} else {
				sColumns = append(sColumns, "key")
				sValues = append(sValues, *o.Key)
			}
		}
		if o.Tags != nil {
			ts := dbutil.StringArray{}
			ts.Load(*o.Tags)
			if len(ts.Strings) > 0 {
				// Need to make sure ALL the tags queried here are available. We assume that the values in database are distinct
				sColumns = append(sColumns, fmt.Sprintf("(SELECT COUNT(json_each.value) FROM json_each(tags) WHERE json_each.value IN (%s))", QQ(len(ts.Strings))))
				for i := range ts.Strings {
					sValues = append(sValues, ts.Strings[i])
				}
				sValues = append(sValues, len(ts.Strings))
			}

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
