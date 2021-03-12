package assets

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// The http verbs to permit in router
var httpVerbs = map[string]bool{
	"GET":    true,
	"POST":   true,
	"PATCH":  true,
	"PUT":    true,
	"DELETE": true,
}

// The permitted prefixes
var routePrefix = map[string]bool{
	"http://":    true,
	"https://":   true,
	"unix://":    true,
	"builtin://": true,
	"run://":     true,
}

func isValidRoute(s string) error {
	ss := strings.Fields(s)
	if len(ss) == 0 {
		return errors.New("Empty route")
	}
	if len(ss) == 1 {
		if !strings.HasPrefix(ss[0], "/") {
			return fmt.Errorf("Route '%s' needs to start with a verb or /", s)
		}
		return nil
	}
	if len(ss) > 2 {
		return fmt.Errorf("Route '%s' must be in format <verb (optional)> <route>", s)
	}
	_, ok := httpVerbs[ss[0]]
	if !ok {
		return fmt.Errorf("Unrecognized http verb '%s' in route '%s'", ss[0], s)
	}
	return nil
}

func isValidTarget(c *Configuration, plugin string, s string) error {
	ss := strings.SplitAfterN(s, "://", 2)
	if len(ss) != 2 {
		return fmt.Errorf("Route target '%s' is missing a prefix", s)
	}
	_, ok := routePrefix[ss[0]]
	if !ok {
		return fmt.Errorf("Route target '%s': unrecognized prefix '%s'", s, ss[0])
	}
	if ss[0] == "run://" {
		// Check to ensure that the given runner was actually defined
		sss := strings.SplitN(ss[1], "/", 2)
		ssss := strings.Split(sss[0], ":")
		if len(ssss) == 0 || len(ssss) > 2 {
			return fmt.Errorf("Route target '%s' invalid", s)
		}
		pname := ssss[0]
		rname := ssss[0]
		if len(ssss) == 1 {
			pname = plugin

		} else {
			rname = ssss[1]
		}

		p, ok := c.Plugins[pname]
		if !ok {
			return fmt.Errorf("Route target '%s' does not exist", s)
		}
		_, ok = (*p).Run[rname]
		if !ok {
			return fmt.Errorf("Route target '%s' does not exist", s)
		}
	}
	return nil
}

func Validate(c *Configuration) error {
	c.RLock()
	defer c.RUnlock()

	if len(c.UserSettingsSchema) > 0 {
		var err error
		c.userSettingsSchema, err = NewSchema(c.UserSettingsSchema)
		if err != nil {
			return err
		}
	}

	for k, v := range c.ObjectTypes {
		err := v.ValidateMeta(nil)
		if err != nil {
			return fmt.Errorf("object %s meta schema invalid: %s", k, err.Error())
		}
	}

	// Make sure all the active plugins have valid configurations
	for _, p := range c.GetActivePlugins() {
		v, ok := c.Plugins[p]
		if !ok {
			return fmt.Errorf("Plugin '%s' config not found", p)
		}
		for conn, v2 := range v.Apps {
			for s, v3 := range v2.Objects {
				if _, ok := c.ObjectTypes[v3.Type]; !ok {
					return fmt.Errorf("[plugin: %s, app: %s, object: %s] unrecognized type (%s)", p, conn, s, v3.Type)
				}
			}
		}
		s, err := NewSchema(v.ConfigSchema)
		if err != nil {
			return err
		}
		if err = s.ValidateAndInsertDefaults(v.Config); err != nil {
			return err
		}
	}

	if c.RunTimeout != nil {
		_, err := time.ParseDuration(*c.RunTimeout)
		if err != nil {
			return errors.New("Invalid exec_timeout")
		}
	}

	// Now make sure all runners are set up correctly
	runners := make(map[string]*JSONSchema)
	for k, v := range c.RunTypes {
		if v.API == nil && k != "exec" && k != "builtin" {
			return fmt.Errorf("RunType '%s' doesn't specify an API target", k)
		} else if v.API != nil {
			if err := isValidTarget(c, "", *v.API); err != nil {
				return err
			}
		}
		s, err := NewSchema(v.ConfigSchema)
		if err != nil {
			return err
		}

		runners[k] = s
	}

	// ...and make sure that all run calls conform to their appropriate schema
	defaultType := "exec"
	for _, pname := range c.GetActivePlugins() {
		p := c.Plugins[pname]
		for _, r := range p.Run {
			if r.Type == nil {
				r.Type = &defaultType
			}
			s, ok := runners[*r.Type]
			if !ok {
				return fmt.Errorf("Unrecognized run type %s", *r.Type)
			}

			if err := s.ValidateAndInsertDefaults(r.Config); err != nil {
				return err
			}
		}
	}

	// Ensure that all routes use permitted verbs and start with permitted route prefix, and validate schemas there too
	for pname, p := range c.Plugins {
		if p.Routes != nil {
			for k, v := range *p.Routes {
				if err := isValidRoute(k); err != nil {
					return err
				}
				if err := isValidTarget(c, pname, v); err != nil {
					return err
				}
			}
		}
		if len(p.UserSettingsSchema) > 0 {
			var err error
			p.userSettingsSchema, err = NewSchema(p.UserSettingsSchema)
			if err != nil {
				return err
			}
		}

		for _, e := range p.On {
			if e.Post == nil {
				return errors.New("'on' must have post specified")
			}
			if err := isValidTarget(c, pname, *e.Post); err != nil {
				return err
			}
		}
		for _, app := range p.Apps {
			for _, e := range app.On {
				if e.Post == nil {
					return errors.New("'on' must have post specified")
				}
				if err := isValidTarget(c, pname, *e.Post); err != nil {
					return err
				}
			}
			for _, s := range app.Objects {
				for _, e := range s.On {
					if e.Post == nil {
						return errors.New("'on' must have post specified")
					}
					if err := isValidTarget(c, pname, *e.Post); err != nil {
						return err
					}
				}
			}

		}
	}
	for _, s := range c.ObjectTypes {
		if s.Routes != nil {
			for k, v := range *s.Routes {
				if err := isValidRoute(k); err != nil {
					return err
				}
				if err := isValidTarget(c, "", v); err != nil {
					return err
				}
			}
		}
		if s.MetaSchema != nil {
			// Can't actually use the schema value here
			_, err := NewSchema(*s.MetaSchema)
			if err != nil {
				return err
			}

		}

	}

	// Finally, set the URL if it isn't set
	if c.URL == nil || *c.URL == "" {
		if c.Port != nil {
			// If port is not set, it means we're testing
			myurl := fmt.Sprintf("http://%s:%d", GetOutboundIP(), *c.Port)
			c.URL = &myurl
		} else {
			testurl := "http://localhost"
			c.URL = &testurl
		}
	}
	if strings.HasSuffix(*c.URL, "/") {
		noslash := (*c.URL)[:len(*c.URL)-1]
		c.URL = &noslash
	}

	return nil
}
