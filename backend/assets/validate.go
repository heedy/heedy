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

	for k, v := range c.SourceTypes {
		err := v.ValidateMeta(nil)
		if err != nil {
			return fmt.Errorf("source %s meta schema invalid: %s", k, err.Error())
		}
	}

	for p, v := range c.Plugins {
		for conn, v2 := range v.Apps {
			for s, v3 := range v2.Sources {
				if _, ok := c.SourceTypes[v3.Type]; !ok {
					return fmt.Errorf("[plugin: %s, app: %s, source: %s] unrecognized type (%s)", p, conn, s, v3.Type)
				}
			}
		}
		s, err := NewSchema(v.SettingsSchema)
		if err != nil {
			return err
		}
		if err = s.ValidateWithDefaults(v.Settings); err != nil {
			return err
		}
	}

	// Make sure all the active plugins have an associated configuration
	for _, ap := range c.GetActivePlugins() {
		if _, ok := c.Plugins[ap]; !ok {
			return fmt.Errorf("Plugin '%s' config not found", ap)
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
		s, err := NewSchema(v.Schema)
		if err != nil {
			return err
		}
		runners[k] = s
	}

	// ...and make sure that all run calls conform to their appropriate schema
	defaultType := "exec"
	for _, p := range c.Plugins {
		for _, r := range p.Run {
			if r.Type == nil {
				r.Type = &defaultType
			}
			s, ok := runners[*r.Type]
			if !ok {
				return fmt.Errorf("Unrecognized run type %s", *r.Type)
			}

			if err := s.ValidateWithDefaults(r.Settings); err != nil {
				return err
			}
		}
	}

	// Ensure that all routes use permitted verbs and start with permitted route prefix
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
			for _, s := range app.Sources {
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
	for _, s := range c.SourceTypes {
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
	}
	for rname, r := range c.RunTypes {
		if r.API == nil {
			return fmt.Errorf("RunType '%s' doesn't specify an API target", rname)
		}
		if err := isValidTarget(c, "", *r.API); err != nil {
			return err
		}
	}

	return nil
}
