package assets

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/xeipuuv/gojsonschema"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

type Event struct {
	Event string  `hcl:"event,label" json:"-"`
	Post  *string `hcl:"post" json:"post,omitempty"`
}

func (e *Event) Validate() error {
	if e.Post == nil {
		return errors.New("'on' must have post specified")
	}
	return nil
}

// Source represents a source that is to be auto-created inside a connection on behalf of a plugin
type Source struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Description *string                 `json:"description,omitempty"`
	Icon        *string                 `json:"icon,omitempty"`
	Meta        *map[string]interface{} `json:"meta,omitempty"`
	Scopes      *[]string               `json:"scopes,omitempty"`

	On map[string]*Event `hcl:"on,block" json:"on,omitempty"`
}

// Connection represents a connection that is to be created on behalf of a plugin
type Connection struct {
	Name string `json:"name"`

	Description *string   `json:"description,omitempty" hcl:"description"`
	Icon        *string   `json:"icon,omitempty" hcl:"icon"`
	AccessToken *bool     `json:"access_token,omitempty" hcl:"access_token"`
	Scopes      *[]string `json:"scopes,omitempty" hcl:"scopes"`
	Type        *string   `json:"type" hcl:"type"`
	Enabled     *bool     `json:"enabled,omitempty" hcl:"enabled"`
	Readonly    *[]string `json:"readonly,omitempty" hcl:"readonly"`

	Settings       *map[string]interface{} `json:"settings,omitempty"`
	SettingsSchema *map[string]interface{} `json:"settings_schema,omitempty"`

	Sources map[string]*Source `json:"sources,omitempty"`

	On map[string]*Event `hcl:"on,block" json:"on,omitempty"`
}

type Exec struct {
	Type      *string   `hcl:"type" json:"enabled,omitempty"`
	Enabled   *bool     `hcl:"enabled" json:"enabled,omitempty"`
	Cron      *string   `hcl:"cron" json:"cron,omitempty"`
	KeepAlive *bool     `hcl:"keepalive" json:"keepalive,omitempty"`
	Cmd       *[]string `hcl:"cmd" json:"cmd,omitempty"`
	Endpoint  *string   `hcl:"endpoint" json:"endpoint,omitempty"`
}

type Plugin struct {
	Version     *string `hcl:"version" json:"version,omitempty"`
	Description *string `hcl:"description" json:"description,omitempty"`
	Icon        *string `hcl:"icon" json:"icon,omitempty"`
	Readme      *string `hcl:"readme" json:"readme,omitempty"`
	Homepage    *string `hcl:"homepage" json:"homepage,omitempty"`
	License     *string `hcl:"license" json:"license,omitempty"`

	Frontend *string            `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`
	Routes   *map[string]string `json:"routes,omitempty"`
	Events   *map[string]string `json:"events,omitempty"`

	On map[string]*Event `hcl:"on,block" json:"on,omitempty"`

	Run      map[string]*Exec    `json:"run,omitempty"`
	Settings map[string]*Setting `json:"settings,omitempty"`

	Connections map[string]*Connection `json:"connections,omitempty"`
}

func (p *Plugin) Copy() *Plugin {
	np := *p
	np.Run = make(map[string]*Exec)
	np.Settings = make(map[string]*Setting)

	for ekey, eval := range p.Run {
		newe := *eval
		np.Run[ekey] = &newe
	}
	for skey, sval := range p.Settings {
		news := *sval
		np.Settings[skey] = &news
	}

	return &np
}

type SourceType struct {
	Frontend *string            `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`
	Routes   *map[string]string `json:"routes,omitempty" hcl:"routes" cty:"routes"`

	Meta *map[string]interface{} `json:"meta,omitempty"`

	Scopes *map[string]string `json:"scopes,omitempty" hcl:"scopes" cty:"scopes"`

	metaSchema *gojsonschema.Schema
	metaObj    map[string]interface{}
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

	return snew
}

// ValidateMeta checks the given metadata is valid
func (s *SourceType) ValidateMeta(meta *map[string]interface{}) (err error) {
	if s.metaSchema == nil {
		objectMap := make(map[string]interface{})
		objectMap["type"] = "object"
		objectMap["additionalProperties"] = false

		if s.Meta != nil {
			if v, ok := (*s.Meta)["type"]; ok {
				if v != "object" {
					return errors.New("Meta schema type must be object")
				}
				objectMap = *s.Meta
			} else {
				propMap := make(map[string]interface{})
				for k, v := range *s.Meta {
					if k == "additionalProperties" || k == "required" {
						objectMap[k] = v
					} else {
						propMap[k] = v
					}
				}
				objectMap["properties"] = propMap
			}
		}
		s.metaObj = objectMap

		// objectMap is now the schema
		s.metaSchema, err = gojsonschema.NewSchema(gojsonschema.NewGoLoader(objectMap))
		if err != nil {
			s.metaSchema = nil
			return err
		}
	}
	if meta != nil {
		// Validate the schema
		res, err := (*s.metaSchema).Validate(gojsonschema.NewGoLoader(meta))
		if err != nil {
			return err
		}
		if !res.Valid() {
			return errors.New(res.Errors()[0].String())
		}
	}

	return nil
}

// ValidateMetaWithDefaults takes a meta value, and adds any required defaults to the root object
// if a default is provided.
func (s *SourceType) ValidateMetaWithDefaults(meta map[string]interface{}) (err error) {
	err = s.ValidateMeta(&meta)
	if err != nil {
		// If there was an issue, we check if there are defaults in the schema for required values
		// that we can set here
		v, ok := s.metaObj["required"]
		if !ok {
			return err
		}
		va, ok := v.([]interface{})
		if !ok {
			return err
		}

		propObji, ok := s.metaObj["properties"]
		if !ok {
			return err
		}
		propObj, ok := propObji.(map[string]interface{})
		if !ok {
			return err
		}

		updated := false
		for _, vav := range va {
			vavs, ok := vav.(string)
			if !ok {
				return err
			}
			_, ok = meta[vavs]
			if !ok {
				// The meta doesn't have the required value. Check if a default is set
				mov, ok := propObj[vavs]
				if !ok {
					return err
				}
				movm, ok := mov.(map[string]interface{})
				if !ok {
					return err
				}
				defaultval, ok := movm["default"]
				if !ok {
					return err
				}
				meta[vavs] = defaultval
				updated = true
			}
		}
		if updated {
			err = s.ValidateMeta(&meta)
		}
	}
	return err
}

type Configuration struct {
	sync.RWMutex

	SiteURL        *string   `hcl:"site_url" json:"site_url,omitempty"`
	Host           *string   `hcl:"host" json:"host,omitempty"`
	Port           *uint16   `hcl:"port" json:"port,omitempty"`
	ActivePlugins  *[]string `hcl:"plugins" json:"plugins,omitempty"`
	AdminUsers     *[]string `hcl:"admin_users" json:"admin_users,omitempty"`
	ForbiddenUsers *[]string `hcl:"forbidden_users" json:"forbidden_users,omitempty"`

	Language         *string `hcl:"language" json:"language,omitempty"`
	FallbackLanguage *string `hcl:"fallback_language" json:"fallback_language,omitempty"`

	SQL *string `hcl:"sql" json:"sql,omitempty"`

	Frontend *string `json:"frontend"`

	RunTimeout *string `json:"run_timeout,omitempty"`

	Scopes              *map[string]string `json:"scopes,omitempty" hcl:"scopes"`
	NewConnectionScopes *[]string          `json:"new_connection_scopes,omitempty" hcl:"new_connection_scopes"`

	SourceTypes map[string]SourceType `json:"source" hcl:"source_types"`

	RequestBodyByteLimit *int64 `hcl:"request_body_byte_limit" json:"request_body_byte_limit,omitempty"`
	AllowPublicWebsocket *bool  `hcl:"allow_public_websocket" json:"allow_public_websocket,omitempty"`

	Plugins map[string]*Plugin `json:"plugin,omitempty"`

	LogLevel *string `json:"log_level" hcl:"log_level"`
	LogFile  *string `json:"log_file" hcl:"log_file"`

	// The verbose option is not possible to set in config, it is passed as an arg. It is only here so that it is passed to plugins
	Verbose bool `json:"verbose"`
}

func (c *Configuration) Validate() error {
	c.RLock()
	defer c.RUnlock()
	if c.SQL == nil {
		return fmt.Errorf("No SQL database was specified")
	}

	for k, v := range c.SourceTypes {
		err := v.ValidateMeta(nil)
		if err != nil {
			return fmt.Errorf("source %s meta schema invalid: %s", k, err.Error())
		}
	}

	for p, v := range c.Plugins {
		for conn, v2 := range v.Connections {
			for s, v3 := range v2.Sources {
				if _, ok := c.SourceTypes[v3.Type]; !ok {
					return fmt.Errorf("[plugin: %s, connection: %s, source: %s] unrecognized type (%s)", p, conn, s, v3.Type)
				}
			}
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
	}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Run:         make(map[string]*Exec),
		Settings:    make(map[string]*Setting),
		Connections: make(map[string]*Connection),
		On:          make(map[string]*Event),
	}
}

func NewConnection() *Connection {
	return &Connection{
		Sources: make(map[string]*Source),
		On:      make(map[string]*Event),
	}
}
func NewSource() *Source {
	return &Source{
		On: make(map[string]*Event),
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

	// Now go into the plugins, and continue the good work
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
			for execName, oexecValue := range oplugin.Run {
				bexecValue, ok := bplugin.Run[execName]
				if !ok {
					bplugin.Run[execName] = oexecValue
				} else {
					CopyStructIfPtrSet(bexecValue, oexecValue)
				}
			}
			for oName, oV := range oplugin.On {
				bV, ok := bplugin.On[oName]
				if !ok {
					bplugin.On[oName] = oV
				} else {
					CopyStructIfPtrSet(bV, oV)
				}
			}

			for cName, ocValue := range oplugin.Connections {
				bcValue, ok := bplugin.Connections[cName]
				if !ok {
					bplugin.Connections[cName] = ocValue
				} else {
					for oName, oV := range ocValue.On {
						bV, ok := bcValue.On[oName]
						if !ok {
							bcValue.On[oName] = oV
						} else {
							CopyStructIfPtrSet(bV, oV)
						}
					}
					CopyStructIfPtrSet(bcValue, ocValue)
					for sName, sValue := range ocValue.Sources {
						bsValue, ok := bcValue.Sources[sName]
						if !ok {
							bcValue.Sources[sName] = sValue
						} else {
							for oName, oV := range sValue.On {
								bV, ok := bsValue.On[oName]
								if !ok {
									bsValue.On[oName] = oV
								} else {
									CopyStructIfPtrSet(bV, oV)
								}
							}
							CopyStructIfPtrSet(bsValue, sValue)
						}
					}
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

	if overlay.Verbose {
		base.Verbose = true
	}

	return base

}

// LoadConfigFile loads configuration from file
func LoadConfigFile(filename string) (*Configuration, error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return LoadConfigBytes(src, filename)
}

// LoadConfigBytes loads the configuration from bytes
func LoadConfigBytes(src []byte, filename string) (*Configuration, error) {
	f, diag := hclsyntax.ParseConfig(src, filename, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diag != nil {
		return nil, diag
	}

	return loadConfigFromHcl(f, filename)
}
