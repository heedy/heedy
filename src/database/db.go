package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Group struct {
	ID          string  `json:"id"`
	Name        *string `json:"name"`
	FullName    *string `json:"fullname"`
	Description *string `json:"description"`
	Owner       *string `json:"owner"`
	Icon        *string `json:"icon"`
}

type User struct {
	Group
	Password string `json:"password,omitempty"`
}

type Device struct {
	Group
	APIKey *string `json:"apikey,omitempty"`
}

type StreamPermissions struct {
	Target string `json:"target"`
	Actor  string `json:"actor"`

	StreamRead   bool `json:"stream_read"`
	StreamWrite  bool `json:"stream_write"`
	StreamDelete bool `json:"stream_delete"`

	DataRead    bool `json:"data_read"`
	DataWrite   bool `json:"data_write"`
	DataRemove  bool `json:"data_remove"`
	ActionWrite bool `json:"action_write"`
}

type GroupPermissions struct {
	StreamPermissions

	GroupRead   bool `json:"group_read"`
	GroupWrite  bool `json:"group_write"`
	GroupDelete bool `json:"group_delete"`

	AddStream bool `json:"add_stream"`
	AddChild  bool `json:"add_child"`

	ListStreams  bool `json:"list_streams"`
	ListChildren bool `json:"list_children"`
	ListShared   bool `json:"list_shared"`
}

// DB represents the database. This interface is implemented in many ways:
//	once for admin
//	once for users
//	once for devices
//	once for public
type DB interface {
	CreateUser(u *User, password string) error
	ReadUser(name string) (*User, error)
}

var (
	ErrNotFound        = errors.New("The selected resource was not found")
	ErrNoUpdate        = errors.New("Nothing to update")
	ErrNoPasswordGiven = errors.New("A user cannot have an empty password")
	ErrUserNotFound    = errors.New("User was not found")
	ErrInvalidName     = errors.New("Invalid name")
	ErrInvalidQuery    = errors.New("Invalid query")
)

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

// Ensures that the icon is in a valid format
func ValidIcon(icon string) error {
	if icon == "" {
		return nil
	}
	// We permit special icon prefixes to be used. The first one is material:, which represents material icons
	// that are assumed to be bundled with all applications that display ConnectorDB data. The second is fa: which
	// will represent fontawesome icons in the future
	if strings.HasPrefix(icon, "material:") || strings.HasPrefix(icon, "fa:") {
		if len(icon) > 30 {
			return errors.New("icon name can't be more than 30 characters.")
		}
		return nil
	}
	_, err := base64.URLEncoding.DecodeString(icon)
	return err
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

func groupTable(g *Group) (groupColumns []string, groupValues []interface{}, err error) {
	groupColumns = make([]string, 0)
	groupValues = make([]interface{}, 0)

	if g.Name != nil {
		if err = ValidName(*g.Name); err != nil {
			return
		}
		groupColumns = append(groupColumns, "name")
		groupValues = append(groupValues, *g.Name)
	}

	if g.Description != nil {
		groupColumns = append(groupColumns, "description")
		groupValues = append(groupValues, *g.Description)
	}
	if g.Icon != nil {
		if err = ValidIcon(*g.Icon); err != nil {
			return
		}
		groupColumns = append(groupColumns, "icon")
		groupValues = append(groupValues, *g.Icon)
	}
	if g.FullName != nil {
		groupColumns = append(groupColumns, "fullname")
		groupValues = append(groupValues, *g.FullName)
	}
	if g.Owner != nil {
		if err = ValidName(*g.Owner); err != nil {
			return
		}
		groupColumns = append(groupColumns, "owner")
		groupValues = append(groupValues, *g.Owner)
	}
	return
}

func userTable(u *User) (groupColumns []string, groupValues []interface{}, userColumns []string, userValues []interface{}, err error) {
	groupColumns, groupValues, err = groupTable(&u.Group)
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

// Insert the right amount of question marks for the given query
func qQ(size int) string {
	s := strings.Repeat("?,", size)
	return s[:len(s)-1]
}

func userCreateQuery(u *User) (string, []interface{}, string, []interface{}, error) {
	if u.Name == nil {
		return "", nil, "", nil, ErrInvalidName
	}
	groupColumns, groupValues, userColumns, userValues, err := userTable(u)
	if err != nil {
		return "", nil, "", nil, err
	}

	// Now add the name of the user to group values and columns
	groupColumns = append(groupColumns, "id", "name")
	groupValues = append(groupValues, *u.Name, *u.Name)

	userColumns = append(userColumns, "id")
	userValues = append(userValues, u.Name)
	return strings.Join(groupColumns, ","), groupValues, strings.Join(userColumns, ","), userValues, err
}

func userUpdateQuery(u *User) (string, []interface{}, string, []interface{}, error) {
	groupColumns, groupValues, userColumns, userValues, err := userTable(u)
	if err != nil {
		return "", nil, "", nil, err
	}

	return strings.Join(groupColumns, "=?,") + "=?", groupValues, strings.Join(userColumns, "=?,") + "=?", userValues, err
}

func groupCreateQuery(g *Group) (string, []interface{}, error) {
	if g.Name == nil {
		return "", nil, ErrInvalidName
	}
	groupColumns, groupValues, err := groupTable(g)
	if err != nil {
		return "", nil, err
	}

	// Since we are creating the group, we also set up the name and id of the group
	// We guarantee that ID is last element
	groupColumns = append(groupColumns, "name", "id")
	groupValues = append(groupValues, g.Name, uuid.New().String())

	return strings.Join(groupColumns, ","), groupValues, nil

}

func groupUpdateQuery(g *Group) (string, []interface{}, error) {
	groupColumns, groupValues, err := groupTable(g)
	return strings.Join(groupColumns, "=?,") + "=?", groupValues, err
}
