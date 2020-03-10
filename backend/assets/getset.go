package assets

import (
	"fmt"
	"time"
)

func (c *Configuration) GetRequestBodyByteLimit() int64 {
	c.RLock()
	defer c.RUnlock()
	if c.RequestBodyByteLimit != nil {
		return *c.RequestBodyByteLimit
	}
	return 0
}

func (c *Configuration) GetHost() string {
	c.RLock()
	defer c.RUnlock()
	if c.Host != nil {
		return *c.Host
	}
	return ""
}

func (c *Configuration) GetPort() uint16 {
	c.RLock()
	defer c.RUnlock()
	if c.Port != nil {
		return *c.Port
	}
	return 0
}

func (c *Configuration) GetActivePlugins() []string {
	c.RLock()
	defer c.RUnlock()
	if c.ActivePlugins == nil {
		return []string{}
	}
	return *c.ActivePlugins
}

// UserIsAdmin checks if the given user is an admin
func (c *Configuration) UserIsAdmin(username string) bool {
	c.RLock()
	defer c.RUnlock()
	if c.AdminUsers == nil {
		return false
	}
	for _, v := range *c.AdminUsers {
		if v == username {
			return true
		}
	}
	return false
}

// GetObjectType returns the given object type
func (c *Configuration) GetObjectType(objecttype string) (*ObjectType, bool) {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.ObjectTypes[objecttype]
	return &s, ok
}

// ValidateObjectMeta makes sure that objects have valid metadata
func (c *Configuration) ValidateObjectMeta(objecttype string, meta *map[string]interface{}) error {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.ObjectTypes[objecttype]
	if !ok {
		return fmt.Errorf("bad_request: invalid object type '%s'", objecttype)
	}
	return s.ValidateMeta(meta)
}

// ValidateObjectMetaUpdate makes sure that objects have valid metadata update queries
func (c *Configuration) ValidateObjectMetaUpdate(objecttype string, meta map[string]interface{}) error {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.ObjectTypes[objecttype]
	if !ok {
		return fmt.Errorf("bad_request: invalid object type '%s'", objecttype)
	}
	return s.ValidateMetaUpdate(meta)
}

// ValidateObjectMetaWithDefaults validates the object, additionally setting required values to defaults
func (c *Configuration) ValidateObjectMetaWithDefaults(objecttype string, meta map[string]interface{}) error {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.ObjectTypes[objecttype]
	if !ok {
		return fmt.Errorf("bad_request: invalid object type '%s'", objecttype)
	}
	return s.ValidateMetaWithDefaults(meta)
}

// GetObjectScope returns the map of scope
func (c *Configuration) GetObjectScope(objecttype string) (map[string]string, error) {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.ObjectTypes[objecttype]
	if !ok {
		return nil, fmt.Errorf("bad_request: invalid object type '%s'", objecttype)
	}
	if s.Scope == nil {
		return make(map[string]string), nil
	}
	return *s.Scope, nil
}

// GetRunTimeout gets timeout for exec
func (c *Configuration) GetRunTimeout() time.Duration {
	c.RLock()
	defer c.RUnlock()
	if c.RunTimeout != nil {
		d, err := time.ParseDuration(*c.RunTimeout)
		if err != nil {
			return d
		}
	}
	d, _ := time.ParseDuration(("5s"))
	return d
}
