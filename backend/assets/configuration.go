package assets

import (
	"errors"
	"io/ioutil"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

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

// Source represents a source that is to be auto-created inside a app on behalf of a plugin
type Source struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Description *string                 `json:"description,omitempty"`
	Icon        *string                 `json:"icon,omitempty"`
	Meta        *map[string]interface{} `json:"meta,omitempty"`
	Scopes      *[]string               `json:"scopes,omitempty"`

	AutoCreate *bool `json:"auto_create,omitempty" hcl:"auto_create"`

	On map[string]*Event `hcl:"on,block" json:"on,omitempty"`
}

// App represents a app that is to be created on behalf of a plugin
type App struct {
	Name string `json:"name"`

	AutoCreate  *bool `json:"auto_create,omitempty" hcl:"auto_create"`
	Unique      *bool `json:"unique,omitempty" hcl:"unique"`
	AccessToken *bool `json:"access_token,omitempty" hcl:"access_token"`

	Description *string   `json:"description,omitempty" hcl:"description"`
	Icon        *string   `json:"icon,omitempty" hcl:"icon"`
	Scopes      *[]string `json:"scopes,omitempty" hcl:"scopes"`
	Type        *string   `json:"type" hcl:"type"`
	Enabled     *bool     `json:"enabled,omitempty" hcl:"enabled"`
	Readonly    *[]string `json:"readonly,omitempty" hcl:"readonly"`

	Settings       *map[string]interface{} `json:"settings,omitempty"`
	SettingsSchema *map[string]interface{} `json:"settings_schema,omitempty"`

	Sources map[string]*Source `json:"sources,omitempty"`

	On map[string]*Event `hcl:"on,block" json:"on,omitempty"`
}
type Run struct {
	Type     *string                `hcl:"type" json:"type,omitempty"`
	Enabled  *bool                  `hcl:"enabled" json:"enabled,omitempty"`
	Cron     *string                `hcl:"cron" json:"cron,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
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

	Run            map[string]Run         `json:"run,omitempty"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	SettingsSchema map[string]interface{} `json:"settings_schema,omitempty"`

	Apps map[string]*App `json:"apps,omitempty"`
}

func (p *Plugin) Copy() *Plugin {
	np := *p
	np.Run = make(map[string]Run)
	np.Settings = make(map[string]interface{})
	np.SettingsSchema = make(map[string]interface{})

	for ekey, eval := range p.Run {
		newrun := Run{
			Settings: make(map[string]interface{}),
		}
		CopyStructIfPtrSet(&newrun, &eval)
		for k, v := range eval.Settings {
			newrun.Settings[k] = v
		}

		np.Run[ekey] = newrun
	}
	for skey, sval := range p.Settings {
		np.Settings[skey] = sval
	}
	for skey, sval := range p.SettingsSchema {
		np.SettingsSchema[skey] = sval
	}

	return &np
}

type SourceType struct {
	Frontend *string            `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`
	Routes   *map[string]string `json:"routes,omitempty" hcl:"routes" cty:"routes"`

	Meta *map[string]interface{} `json:"meta,omitempty"`

	Scopes *map[string]string `json:"scopes,omitempty" hcl:"scopes" cty:"scopes"`

	metaSchema *JSONSchema
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
		if s.Meta != nil {
			s.metaSchema, err = NewSchema(*s.Meta)
		} else {
			s.metaSchema, err = NewSchema(make(map[string]interface{}))
		}
		if err != nil {
			return
		}
	}
	if meta != nil {
		// Validate the schema
		return s.metaSchema.Validate(*meta)
	}

	return nil
}

// ValidateMetaWithDefaults takes a meta value, and adds any required defaults to the root object
// if a default is provided.
func (s *SourceType) ValidateMetaWithDefaults(meta map[string]interface{}) (err error) {
	if s.metaSchema == nil {
		if s.Meta != nil {
			s.metaSchema, err = NewSchema(*s.Meta)
		} else {
			s.metaSchema, err = NewSchema(make(map[string]interface{}))
		}
		if err != nil {
			return
		}
	}
	return s.metaSchema.ValidateWithDefaults(meta)
}

type RunType struct {
	Schema map[string]interface{} `json:"schema,omitempty" hcl:"schema" cty:"schema"`
	API    *string                `json:"api,omitempty" hcl:"api" cty:"api"`
}

func (r *RunType) Copy() RunType {
	rnew := RunType{}
	CopyStructIfPtrSet(&rnew, r)

	rnew.Schema = r.Schema
	return rnew
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

	Frontend *string `json:"frontend,omitempty"`

	RunTimeout *string `json:"run_timeout,omitempty"`

	Scopes       *map[string]string `json:"scopes,omitempty" hcl:"scopes"`
	NewAppScopes *[]string          `json:"new_app_scopes,omitempty" hcl:"new_app_scopes"`

	SourceTypes map[string]SourceType `json:"source,omitempty" hcl:"source_types"`
	RunTypes    map[string]RunType    `json:"runtype,omitempty"`

	RequestBodyByteLimit *int64 `hcl:"request_body_byte_limit" json:"request_body_byte_limit,omitempty"`
	AllowPublicWebsocket *bool  `hcl:"allow_public_websocket" json:"allow_public_websocket,omitempty"`

	Plugins map[string]*Plugin `json:"plugin,omitempty"`

	LogLevel *string `json:"log_level,omitempty" hcl:"log_level"`
	LogFile  *string `json:"log_file,omitempty" hcl:"log_file"`

	// The verbose option is not possible to set in config, it is passed as an arg. It is only here so that it is passed to plugins
	Verbose bool `json:"verbose,omitempty"`
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
		RunTypes:    make(map[string]RunType),
	}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Run:      make(map[string]Run),
		Settings: make(map[string]interface{}),
		Apps:     make(map[string]*App),
		On:       make(map[string]*Event),
	}
}

func NewApp() *App {
	return &App{
		Sources: make(map[string]*Source),
		On:      make(map[string]*Event),
	}
}
func NewSource() *Source {
	return &Source{
		On: make(map[string]*Event),
	}
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
	overlay.NewAppScopes = MergeStringArrays(base.NewAppScopes, overlay.NewAppScopes)
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

	for k, v := range overlay.RunTypes {
		bv, ok := base.RunTypes[k]
		if ok {
			CopyStructIfPtrSet(&bv, &v)
			if len(v.Schema) > 0 {
				bv.Schema = v.Schema
			}
			base.RunTypes[k] = bv
		} else {
			base.RunTypes[k] = v
		}
	}

	// Now go into the plugins, and continue the good work
	for pluginName, oplugin := range overlay.Plugins {
		bplugin, ok := base.Plugins[pluginName]
		if !ok {
			// We take the overlay's plugin wholesale
			base.Plugins[pluginName] = oplugin

		} else {
			// Need to continue settings merge into the children
			CopyStructIfPtrSet(bplugin, oplugin)

			// Exec jobs
			for execName, oexecValue := range oplugin.Run {
				bexecValue, ok := bplugin.Run[execName]
				if !ok {
					bplugin.Run[execName] = oexecValue
				} else {
					CopyStructIfPtrSet(&bexecValue, &oexecValue)
					for rsn, rsv := range oexecValue.Settings {
						bexecValue.Settings[rsn] = rsv
					}
					bplugin.Run[execName] = bexecValue
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

			for cName, ocValue := range oplugin.Apps {
				bcValue, ok := bplugin.Apps[cName]
				if !ok {
					bplugin.Apps[cName] = ocValue
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
				bplugin.Settings[settingName] = osettingValue
			}
			// Schema copy
			if len(oplugin.SettingsSchema) > 0 {
				bplugin.SettingsSchema = oplugin.SettingsSchema
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
