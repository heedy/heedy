package assets

import (
	"fmt"
	"reflect"
	"strings"
)

type Setting struct {
	Title            *string     `hcl:"title" json:"title,omitempty"`
	Type             *string     `hcl:"type" json:"type,omitempty"`
	Description      *string     `hcl:"description" json:"description,omitempty"`
	Minimum          *float64    `hcl:"minimum" json:"minimum,omitempty"`
	ExclusiveMinimum *float64    `hcl:"exclusiveMinimum" json:"exclusiveMinimum,omitempty"`
	Maximum          *float64    `hcl:"maximum" json:"maximum,omitempty"`
	ExclusiveMaximum *float64    `hcl:"exclusiveMaximum" json:"exclusiveMaximum,omitempty"`
	Items            *Setting    `hcl:"items" json:"items,omitempty"`
	MinItems         *int        `hcl:"minItems" json:"minItems,omitempty"`
	UniqueItems      *bool       `hcl:"uniqueItems" json:"uniqueItems,omitempty"`
	Default          interface{} `json:"default,omitempty"`
	Value            interface{} `json:"value,omitempty"`
}

type Exec struct {
	Enabled   *bool     `hcl:"enabled" json:"enabled,omitempty"`
	Cron      *string   `hcl:"cron" json:"cron,omitempty"`
	KeepAlive *bool     `hcl:"keepalive" json:"keepalive,omitempty"`
	Cmd       *[]string `hcl:"cmd" json:"cmd,omitempty"`
}

type Plugin struct {
	Version     *string            `hcl:"version" json:"version,omitempty"`
	Description *string            `hcl:"description" json:"description,omitempty"`
	Homepage    *string            `hcl:"homepage" json:"homepage,omitempty"`
	License     *string            `hcl:"license" json:"license,omitempty"`
	GRPC        *string            `hcl:"grpc" json:"grpc,omitempty"`
	Routes      *map[string]string `json:"routes,omitempty"`

	//FallbackLanguage *string `hcl:"fallback_language" json:"fallback_language"`

	Exec     map[string]*Exec    `json:"exec,omitempty"`
	Settings map[string]*Setting `json:"settings,omitempty"`
}

func (p *Plugin) Copy() *Plugin {
	np := *p
	np.Exec = make(map[string]*Exec)
	np.Settings = make(map[string]*Setting)

	for ekey, eval := range p.Exec {
		newe := *eval
		np.Exec[ekey] = &newe
	}
	for skey, sval := range p.Settings {
		news := *sval
		np.Settings[skey] = &news
	}

	return &np
}

type MenuItem struct {
	Route *string `json:"route,omitempty" hcl:"route" cty:"route"`
	Icon  *string `json:"icon,omitempty" hcl:"icon" cty:"icon"`
	Text  *string `json:"text,omitempty" hcl:"text" cty:"text"`

	// Description is shown in tooltip
	Description *string `json:"description,omitempty" hcl:"description" cty:"description"`

	// Active is true by default, but can be set to false to disable the route
	Active *bool `json:"active,omitempty" hcl:"active" cty:"active"`
}

type Frontend struct {
	Routes map[string]string   `json:"routes" hcl:"routes"`
	Menu   map[string]MenuItem `json:"menu" hcl:"menu"`

	PublicRoutes map[string]string   `json:"public_routes" hcl:"public_routes"`
	PublicMenu   map[string]MenuItem `json:"public_menu" hcl:"public_menu"`
}

func NewFrontend() Frontend {
	return Frontend{
		Routes:       make(map[string]string),
		PublicRoutes: make(map[string]string),
		Menu:         make(map[string]MenuItem),
		PublicMenu:   make(map[string]MenuItem),
	}
}

func (a *Frontend) Copy() Frontend {
	na := NewFrontend()

	for ak, av := range a.Routes {
		na.Routes[ak] = av
	}
	for ak, av := range a.PublicRoutes {
		na.PublicRoutes[ak] = av
	}

	for ak, av := range a.Menu {
		na.Menu[ak] = av
	}
	for ak, av := range a.PublicMenu {
		na.PublicMenu[ak] = av
	}
	return na
}

// SourceTypeFrontend is the frontend info for
type SourceTypeFrontend struct {
	Name   *string            `json:"name,omitempty" hcl:"name" cty:"name"`
	Icon   *string            `json:"icon,omitempty" hcl:"icon" cty:"icon"`
	Routes *map[string]string `json:"routes,omitempty" hcl:"routes" cty:"routes"`
}

func (f *SourceTypeFrontend) Copy() *SourceTypeFrontend {
	fnew := &SourceTypeFrontend{}
	CopyStructIfPtrSet(fnew, f)
	if f.Routes != nil {
		newRoutes := make(map[string]string)
		for k, v := range *(f.Routes) {
			newRoutes[k] = v
		}
		fnew.Routes = &newRoutes
	}
	return fnew
}

type SourceType struct {
	Frontend *SourceTypeFrontend `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`

	Meta   *map[string]interface{} `json:"meta,omitempty"`
	Routes *map[string]string      `json:"routes,omitempty" hcl:"routes" cty:"routes"`
	Scopes *map[string]string      `json:"scopes,omitempty" hcl:"scopes" cty:"scopes"`
}

func (s *SourceType) Copy() SourceType {
	snew := SourceType{}
	CopyStructIfPtrSet(&snew, s)
	if s.Routes != nil {
		newRoutes := make(map[string]string)
		for k, v := range *(s.Routes) {
			newRoutes[k] = v
		}
		snew.Routes = &newRoutes
	}
	if s.Frontend != nil {
		snew.Frontend = s.Frontend.Copy()
	}

	return snew
}

type Configuration struct {
	SiteURL        *string   `hcl:"site_url" json:"site_url,omitempty"`
	Host           *string   `hcl:"host" json:"host,omitempty"`
	Port           *uint16   `hcl:"port" json:"port,omitempty"`
	ActivePlugins  *[]string `hcl:"plugins" json:"plugins,omitempty"`
	AdminUsers     *[]string `hcl:"admin_users" json:"admin_users,omitempty"`
	ForbiddenUsers *[]string `hcl:"forbidden_users" json:"forbidden_users,omitempty"`

	Language         *string `hcl:"language" json:"language,omitempty"`
	FallbackLanguage *string `hcl:"fallback_language" json:"fallback_language,omitempty"`

	SQL *string `hcl:"sql" json:"sql,omitempty"`

	Frontend Frontend `json:"frontend"`

	ExecTimeout *string `json:"exec_timeout,omitempty"`

	Scopes              *map[string]string `json:"scopes,omitempty" hcl:"scopes"`
	NewConnectionScopes *[]string          `json:"new_connection_scopes,omitempty" hcl:"new_connection_scopes"`

	SourceTypes map[string]SourceType `json:"source" hcl:"source_types"`

	RequestBodyByteLimit *int64 `hcl:"request_body_byte_limit" json:"request_body_byte_limit,omitempty"`

	Plugins map[string]*Plugin `json:"plugin,omitempty"`
}

func (c *Configuration) Validate() error {
	if c.SQL == nil {
		return fmt.Errorf("No SQL database was specified")
	}
	return nil
}

func copyStringArrayPtr(s *[]string) *[]string {
	if s == nil {
		return s
	}
	ns := make([]string, len(*s))
	copy(ns, *s)
	return &ns
}
func (c *Configuration) Copy() *Configuration {
	nc := *c

	nc.Frontend = c.Frontend.Copy()

	nc.Plugins = make(map[string]*Plugin)

	for pkey, pval := range c.Plugins {
		nc.Plugins[pkey] = pval.Copy()
	}

	nc.SourceTypes = make(map[string]SourceType)
	for k, v := range c.SourceTypes {
		nc.SourceTypes[k] = v.Copy()
	}

	return &nc

}

func NewConfiguration() *Configuration {
	return &Configuration{
		Plugins:     make(map[string]*Plugin),
		SourceTypes: make(map[string]SourceType),
		Frontend:    NewFrontend(),
	}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Exec:     make(map[string]*Exec),
		Settings: make(map[string]*Setting),
	}
}

// MergeStringArrays allows merging arrays of strings, with the result having each element
// at most once, and special prefix of + being ignored, and - allowing removal from array
func MergeStringArrays(base *[]string, overlay *[]string) *[]string {
	if base == nil {
		return overlay
	}
	if overlay == nil {
		return base
	}

	output := make([]string, 0)
	for _, d := range *base {
		if !strings.HasPrefix(d, "-") {
			if strings.HasPrefix(d, "+") {
				d = d[1:len(d)]
			}

			// Check if the output aready contains it
			contained := false
			for _, bd := range output {
				if bd == d {
					contained = true
					break
				}
			}
			if !contained {
				output = append(output, d)
			}

		}
	}
	for _, d := range *overlay {
		if strings.HasPrefix(d, "-") {
			if len(output) <= 0 {
				break
			}
			d = d[1:len(d)]

			// Remove element if contained
			for j, bd := range output {
				if bd == d {
					if len(output) == j+1 {
						output = output[:len(output)-1]
					} else {
						output[j] = output[len(output)-1]
						output = output[:len(output)-1]
						break
					}

				}
			}
		} else {
			if strings.HasPrefix(d, "+") {
				d = d[1:len(d)]
			}

			// Check if the output aready contains it
			contained := false
			for _, bd := range output {
				if bd == d {
					contained = true
					break
				}
			}
			if !contained {
				output = append(output, d)
			}
		}
	}
	return &output
}

// Merges two configurations together
func MergeConfig(base *Configuration, overlay *Configuration) *Configuration {
	base = base.Copy()
	overlay = overlay.Copy()

	// Copy the scopes to overlay, since they will be replaced with CopyStruct
	if overlay.Scopes != nil && base.Scopes != nil {
		for sk, sv := range *overlay.Scopes {
			(*base.Scopes)[sk] = sv
		}
		overlay.Scopes = base.Scopes
	}
	overlay.NewConnectionScopes = MergeStringArrays(base.NewConnectionScopes, overlay.NewConnectionScopes)
	overlay.ForbiddenUsers = MergeStringArrays(base.ForbiddenUsers, overlay.ForbiddenUsers)

	CopyStructIfPtrSet(base, overlay)

	// Merge the SourceTypes map
	for ak, av := range overlay.SourceTypes {
		cv, ok := base.SourceTypes[ak]
		if ok {
			// Merge the Frontend overlay
			if cv.Frontend != nil && av.Frontend != nil {

				avf := av.Frontend
				// CopyStruct will replace the frontend with the overlay
				av.Frontend = cv.Frontend

				// av.Frontend is the base, and avf is the overlay
				if av.Frontend.Routes != nil && avf.Routes != nil {
					// Need to merge the routes
					for rk, rv := range *(avf.Routes) {
						(*av.Frontend.Routes)[rk] = rv
					}
					avf.Routes = nil // Set it to nil so it isn't copied over
				}
				CopyStructIfPtrSet(av.Frontend, avf)
			}
			// Copy the scopes to av
			if av.Scopes != nil && cv.Scopes != nil {
				for sk, sv := range *av.Scopes {
					(*cv.Scopes)[sk] = sv
				}
				av.Scopes = cv.Scopes
			}
			// Copy the routes to av
			if av.Routes != nil && cv.Routes != nil {
				for rk, rv := range *av.Routes {
					(*cv.Routes)[rk] = rv
				}
				av.Routes = cv.Routes
			}

			// Update only the set values
			CopyStructIfPtrSet(&cv, &av)
			base.SourceTypes[ak] = cv
		} else {
			base.SourceTypes[ak] = av
		}
	}

	// Merge the maps of Frontend
	for ak, av := range overlay.Frontend.Menu {
		cv, ok := base.Frontend.Menu[ak]
		if ok {
			// Update only the set values of menu
			CopyStructIfPtrSet(&cv, &av)
			base.Frontend.Menu[ak] = cv
		} else {
			base.Frontend.Menu[ak] = av
		}
	}
	for ak, av := range overlay.Frontend.PublicMenu {
		cv, ok := base.Frontend.PublicMenu[ak]
		if ok {
			// Update only the set values of menu
			CopyStructIfPtrSet(&cv, &av)
			base.Frontend.PublicMenu[ak] = cv
		} else {
			base.Frontend.PublicMenu[ak] = av
		}
	}
	for ak, av := range overlay.Frontend.Routes {
		base.Frontend.Routes[ak] = av
	}
	for ak, av := range overlay.Frontend.PublicRoutes {
		base.Frontend.PublicRoutes[ak] = av
	}

	// Now go into the maps, and continue the good work
	for pluginName, oplugin := range overlay.Plugins {
		bplugin, ok := base.Plugins[pluginName]
		if !ok {
			// We take the overlay's plugin wholesale
			base.Plugins[pluginName] = oplugin

			// And any setting values automatically become the defaults, because it is assumed
			// that this config file is defining the given plugin
			for _, setting := range oplugin.Settings {
				if setting.Value != nil {
					setting.Default = setting.Value
				}
			}
		} else {
			// Need to continue settings merge into the children
			CopyStructIfPtrSet(bplugin, oplugin)

			// Exec jobs
			for execName, oexecValue := range oplugin.Exec {
				bexecValue, ok := bplugin.Exec[execName]
				if !ok {
					bplugin.Exec[execName] = oexecValue
				} else {
					CopyStructIfPtrSet(bexecValue, oexecValue)
				}
			}

			// Settings copy
			for settingName, osettingValue := range oplugin.Settings {
				bsettingValue, ok := bplugin.Settings[settingName]
				if !ok {
					bplugin.Settings[settingName] = osettingValue
				} else {
					CopyStructIfPtrSet(bsettingValue, osettingValue)

					// CopyStruct won't copy the interface values, since they might not be ptrs
					if reflect.ValueOf(osettingValue.Default).IsValid() {
						bsettingValue.Default = osettingValue.Default
					}
					if reflect.ValueOf(osettingValue.Value).IsValid() {
						bsettingValue.Value = osettingValue.Value
					}
				}
			}
		}
	}

	return base

}

// LoadConfigFile loads configuration from file
func LoadConfigFile(filename string) (*Configuration, error) {

	f, diag := configparser.ParseHCLFile(filename)
	if diag != nil {
		return nil, diag
	}

	return loadConfigFromHcl(f, filename)
}

// LoadConfigBytes loads the configuration from bytes
func LoadConfigBytes(src []byte, filename string) (*Configuration, error) {
	f, diag := configparser.ParseHCL(src, filename)
	if diag != nil {
		return nil, diag
	}

	return loadConfigFromHcl(f, filename)
}
