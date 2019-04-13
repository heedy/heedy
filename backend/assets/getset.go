package assets

func (c *Configuration) GetRequestBodyByteLimit() int64 {
	if c.RequestBodyByteLimit != nil {
		return *c.RequestBodyByteLimit
	}
	return 0
}

func (c *Configuration) GetHost() string {
	if c.Host != nil {
		return *c.Host
	}
	return ""
}

func (c *Configuration) GetPort() uint16 {
	if c.Port != nil {
		return *c.Port
	}
	return 0
}

func (c *Configuration) GetNewConnectionScopes() []string {
	if c.NewConnectionScopes != nil {
		return *c.NewConnectionScopes
	}
	return []string{}
}

// UserIsAdmin checks if the given user is an admin
func (c *Configuration) UserIsAdmin(username string) bool {
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
