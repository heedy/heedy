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

func (c *Configuration) GetNewAppScopes() []string {
	c.RLock()
	defer c.RUnlock()
	if c.NewAppScopes != nil {
		return *c.NewAppScopes
	}
	return []string{}
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

// GetSourceType returns the given source type
func (c *Configuration) GetSourceType(sourcetype string) (*SourceType, bool) {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.SourceTypes[sourcetype]
	return &s, ok
}

// ValidateSourceMeta makes sure that sources have valid metadata
func (c *Configuration) ValidateSourceMeta(sourcetype string, meta *map[string]interface{}) error {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.SourceTypes[sourcetype]
	if !ok {
		return fmt.Errorf("bad_request: invalid source type '%s'", sourcetype)
	}
	return s.ValidateMeta(meta)
}

// ValidateSourceMetaWithDefaults validates the source, additionally setting required values to defaults
func (c *Configuration) ValidateSourceMetaWithDefaults(sourcetype string, meta map[string]interface{}) error {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.SourceTypes[sourcetype]
	if !ok {
		return fmt.Errorf("bad_request: invalid source type '%s'", sourcetype)
	}
	return s.ValidateMetaWithDefaults(meta)
}

// GetSourceScopes returns the map of scopes
func (c *Configuration) GetSourceScopes(sourcetype string) (map[string]string, error) {
	c.RLock()
	defer c.RUnlock()
	s, ok := c.SourceTypes[sourcetype]
	if !ok {
		return nil, fmt.Errorf("bad_request: invalid source type '%s'", sourcetype)
	}
	if s.Scopes == nil {
		return make(map[string]string), nil
	}
	return *s.Scopes, nil
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
