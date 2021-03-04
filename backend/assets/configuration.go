package assets

import (
	"errors"
	"io/ioutil"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Event struct {
	Event  string  `hcl:"event,label" json:"event,omitempty"`
	Type   *string `hcl:"type" json:"type,omitempty"`
	Tags   *string `hcl:"tags" json:"tags,omitempty"`
	Plugin *string `hcl:"plugin" json:"plugin,omitempty"`
	Key    *string `hcl:"key" json:"key,omitempty"`
	Post   *string `hcl:"post" json:"post,omitempty"`
}

func (e *Event) Validate() error {
	if e.Post == nil {
		return errors.New("'on' must have post specified")
	}
	return nil
}

// Object represents a object that is to be auto-created inside a app on behalf of a plugin
type Object struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Description *string                 `json:"description,omitempty"`
	Icon        *string                 `json:"icon,omitempty"`
	Meta        *map[string]interface{} `json:"meta,omitempty"`
	OwnerScope  *string                 `json:"owner_scope,omitempty"`
	Tags        *string                 `json:"tags,omitempty"`

	AutoCreate *bool `json:"auto_create,omitempty" hcl:"auto_create"`

	On []Event `hcl:"on,block" json:"on,omitempty"`
}

// App represents a app that is to be created on behalf of a plugin
type App struct {
	Name string `json:"name"`

	AutoCreate  *bool `json:"auto_create,omitempty" hcl:"auto_create"`
	Unique      *bool `json:"unique,omitempty" hcl:"unique"`
	AccessToken *bool `json:"access_token,omitempty" hcl:"access_token"`

	Description *string   `json:"description,omitempty" hcl:"description"`
	Icon        *string   `json:"icon,omitempty" hcl:"icon"`
	Scope       *string   `json:"scope,omitempty" hcl:"scope"`
	Enabled     *bool     `json:"enabled,omitempty" hcl:"enabled"`
	Readonly    *[]string `json:"readonly,omitempty" hcl:"readonly"`

	Settings       *map[string]interface{} `json:"settings,omitempty"`
	SettingsSchema *map[string]interface{} `json:"settings_schema,omitempty"`

	Objects map[string]*Object `json:"objects,omitempty"`

	On []Event `hcl:"on,block" json:"on,omitempty"`
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
	Homepage    *string `hcl:"homepage" json:"homepage,omitempty"`
	License     *string `hcl:"license" json:"license,omitempty"`

	Frontend *string   `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`
	Preload  *[]string `json:"preload,omitempty" hcl:"preload"`

	Routes *map[string]string `json:"routes,omitempty"`
	Events *map[string]string `json:"events,omitempty"`

	On []Event `hcl:"on,block" json:"on,omitempty"`

	Run               map[string]Run         `json:"run,omitempty"`
	Settings          map[string]interface{} `json:"settings,omitempty"`
	SettingsSchema    map[string]interface{} `json:"settings_schema,omitempty"`
	PreferencesSchema map[string]interface{} `json:"preferences_schema,omitempty"`

	Apps map[string]*App `json:"apps,omitempty"`

	preferencesSchema *JSONSchema
}

func (p *Plugin) Copy() *Plugin {
	np := *p
	np.Run = make(map[string]Run)
	np.Settings = make(map[string]interface{})
	np.PreferencesSchema = make(map[string]interface{})
	np.SettingsSchema = make(map[string]interface{})
	np.On = make([]Event, len(p.On))

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
	for skey, sval := range p.PreferencesSchema {
		np.PreferencesSchema[skey] = sval
	}
	for si, sval := range p.On {
		np.On[si] = sval
	}

	return &np
}

func (p *Plugin) InsertPreferenceDefaults(prefs map[string]interface{}) (err error) {
	if p.preferencesSchema == nil {
		p.preferencesSchema, err = NewSchema(p.PreferencesSchema)
	}
	p.preferencesSchema.InsertDefaults(prefs)
	return
}

func (p *Plugin) ValidatePreferencesUpdate(prefs map[string]interface{}) (err error) {
	if p.preferencesSchema == nil {
		p.preferencesSchema, err = NewSchema(p.PreferencesSchema)
	}
	return p.preferencesSchema.ValidateUpdate(prefs)
}

func (p *Plugin) GetPreferenceSchema() map[string]interface{} {
	if len(p.PreferencesSchema) == 0 {
		return nil
	}
	if p.preferencesSchema == nil {
		var err error
		p.preferencesSchema, err = NewSchema(p.PreferencesSchema)
		if err != nil {
			return nil
		}
	}
	return p.preferencesSchema.Schema
}

type ObjectType struct {
	Frontend *string            `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`
	Routes   *map[string]string `json:"routes,omitempty" hcl:"routes" cty:"routes"`

	MetaSchema *map[string]interface{} `json:"meta_schema,omitempty"`

	Scope *map[string]string `json:"scope,omitempty" hcl:"scope" cty:"scope"`

	metaSchema *JSONSchema
}

func (s *ObjectType) Copy() ObjectType {
	snew := ObjectType{}
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
func (s *ObjectType) ValidateMeta(meta *map[string]interface{}) (err error) {
	if s.metaSchema == nil {
		if s.MetaSchema != nil {
			s.metaSchema, err = NewSchema(*s.MetaSchema)
		} else {
			s.metaSchema, err = NewSchema(make(map[string]interface{}))
		}
		if err != nil {
			return err
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
func (s *ObjectType) ValidateMetaWithDefaults(meta map[string]interface{}) (err error) {
	if s.metaSchema == nil {
		if s.MetaSchema != nil {
			s.metaSchema, err = NewSchema(*s.MetaSchema)
		} else {
			s.metaSchema, err = NewSchema(make(map[string]interface{}))
		}
		if err != nil {
			return
		}
	}
	return s.metaSchema.ValidateWithDefaults(meta)
}

// ValidateMetaUpdate validates an update query
func (s *ObjectType) ValidateMetaUpdate(meta map[string]interface{}) (err error) {
	if s.metaSchema == nil {
		if s.MetaSchema != nil {
			s.metaSchema, err = NewSchema(*s.MetaSchema)
		} else {
			s.metaSchema, err = NewSchema(make(map[string]interface{}))
		}
		if err != nil {
			return
		}
	}
	return s.metaSchema.ValidateUpdate(meta)
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

	URL            *string   `hcl:"url" json:"url,omitempty"`
	Host           *string   `hcl:"host" json:"host,omitempty"`
	Port           *uint16   `hcl:"port" json:"port,omitempty"`
	ActivePlugins  *[]string `hcl:"active_plugins" json:"active_plugins,omitempty"`
	AdminUsers     *[]string `hcl:"admin_users" json:"admin_users,omitempty"`
	ForbiddenUsers *[]string `hcl:"forbidden_users" json:"forbidden_users,omitempty"`

	Language         *string `hcl:"language" json:"language,omitempty"`
	FallbackLanguage *string `hcl:"fallback_language" json:"fallback_language,omitempty"`

	SQL *string `hcl:"sql" json:"sql,omitempty"`

	Frontend *string   `json:"frontend,omitempty"`
	Preload  *[]string `json:"preload,omitempty" hcl:"preload"`

	RunTimeout *string `json:"run_timeout,omitempty"`

	Scope *map[string]string `json:"scope,omitempty" hcl:"scope"`

	ObjectTypes map[string]ObjectType `json:"type,omitempty" hcl:"type"`
	RunTypes    map[string]RunType    `json:"runtype,omitempty"`

	RequestBodyByteLimit *int64 `hcl:"request_body_byte_limit" json:"request_body_byte_limit,omitempty"`
	AllowPublicWebsocket *bool  `hcl:"allow_public_websocket" json:"allow_public_websocket,omitempty"`

	Plugins map[string]*Plugin `json:"plugin,omitempty"`

	LogLevel *string `json:"log_level,omitempty" hcl:"log_level"`
	LogFile  *string `json:"log_file,omitempty" hcl:"log_file"`

	// Schema for the core UI preferences
	PreferencesSchema map[string]interface{} `json:"preferences_schema,omitempty"`

	// The verbose option is not possible to set in config, it is passed as an arg. It is only here so that it is passed to plugins
	Verbose bool `json:"verbose,omitempty"`

	preferencesSchema *JSONSchema
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
	nc := *c // LOCK VALUE COPY is OK, since we don't use the locks until the config is all loaded.

	nc.Plugins = make(map[string]*Plugin)
	nc.PreferencesSchema = make(map[string]interface{})

	for pkey, pval := range c.Plugins {
		nc.Plugins[pkey] = pval.Copy()
	}
	for skey, sval := range c.PreferencesSchema {
		nc.PreferencesSchema[skey] = sval
	}

	nc.ObjectTypes = make(map[string]ObjectType)
	for k, v := range c.ObjectTypes {
		nc.ObjectTypes[k] = v.Copy()
	}

	return &nc

}

func (c *Configuration) InsertPreferenceDefaults(prefs map[string]interface{}) (err error) {
	if c.preferencesSchema == nil {
		c.preferencesSchema, err = NewSchema(c.PreferencesSchema)
	}
	c.preferencesSchema.InsertDefaults(prefs)
	return
}

func (c *Configuration) ValidateHeedyPreferencesUpdate(prefs map[string]interface{}) (err error) {
	if c.preferencesSchema == nil {
		c.preferencesSchema, err = NewSchema(c.PreferencesSchema)
	}
	return c.preferencesSchema.ValidateUpdate(prefs)
}

func (c *Configuration) GetPreferenceSchema() map[string]interface{} {
	if len(c.PreferencesSchema) == 0 {
		return nil
	}
	if c.preferencesSchema == nil {
		var err error
		c.preferencesSchema, err = NewSchema(c.PreferencesSchema)
		if err != nil {
			return nil
		}
	}
	return c.preferencesSchema.Schema
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Plugins:     make(map[string]*Plugin),
		ObjectTypes: make(map[string]ObjectType),
		RunTypes:    make(map[string]RunType),
	}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Run:      make(map[string]Run),
		Settings: make(map[string]interface{}),
		Apps:     make(map[string]*App),
		On:       make([]Event, 0),
	}
}

func NewApp() *App {
	return &App{
		Objects: make(map[string]*Object),
		On:      make([]Event, 0),
	}
}
func NewObject() *Object {
	return &Object{
		On: make([]Event, 0),
	}
}

func MergeMap(to, from map[string]interface{}) {
	for k, v := range from {
		if v == nil {
			delete(to, k)
		} else {
			to[k] = v
		}
	}
}

// Merges two configurations together
func MergeConfig(base *Configuration, overlay *Configuration) *Configuration {
	base = base.Copy()
	overlay = overlay.Copy()

	// Copy the scope to overlay, since they will be replaced with CopyStruct
	if overlay.Scope != nil && base.Scope != nil {
		for sk, sv := range *overlay.Scope {
			(*base.Scope)[sk] = sv
		}
		overlay.Scope = base.Scope
	}
	overlay.ForbiddenUsers = MergeStringArrays(base.ForbiddenUsers, overlay.ForbiddenUsers)
	overlay.Preload = MergeStringArrays(base.Preload, overlay.Preload)

	CopyStructIfPtrSet(base, overlay)

	if len(overlay.PreferencesSchema) > 0 {
		MergeMap(base.PreferencesSchema, overlay.PreferencesSchema)
	}

	// Merge the ObjectTypes map
	for ak, av := range overlay.ObjectTypes {
		cv, ok := base.ObjectTypes[ak]
		if ok {
			// Copy the scope to av
			if av.Scope != nil && cv.Scope != nil {
				for sk, sv := range *av.Scope {
					(*cv.Scope)[sk] = sv
				}
				av.Scope = cv.Scope
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
			base.ObjectTypes[ak] = cv
		} else {
			base.ObjectTypes[ak] = av
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
			oplugin.Preload = MergeStringArrays(bplugin.Preload, oplugin.Preload)
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
			for _, oV := range oplugin.On {
				bplugin.On = append(bplugin.On, oV)
			}

			for cName, ocValue := range oplugin.Apps {
				bcValue, ok := bplugin.Apps[cName]
				if !ok {
					bplugin.Apps[cName] = ocValue
				} else {
					for _, oV := range ocValue.On {
						bcValue.On = append(bcValue.On, oV)
					}
					CopyStructIfPtrSet(bcValue, ocValue)
					for sName, sValue := range ocValue.Objects {
						bsValue, ok := bcValue.Objects[sName]
						if !ok {
							bcValue.Objects[sName] = sValue
						} else {
							for _, oV := range sValue.On {
								bsValue.On = append(bsValue.On, oV)
							}
							CopyStructIfPtrSet(bsValue, sValue)
						}
					}
				}
			}

			// Schema copy
			if len(oplugin.SettingsSchema) > 0 {
				MergeMap(bplugin.SettingsSchema, oplugin.SettingsSchema)
			}
			if len(oplugin.PreferencesSchema) > 0 {
				MergeMap(bplugin.PreferencesSchema, oplugin.PreferencesSchema)
			}

			// Settings copy
			for settingName, osettingValue := range oplugin.Settings {
				// If the setting values are both objects,
				// then merge the object's first level. Otherwise, replace

				v, ok := bplugin.Settings[settingName]
				if ok {
					var v2 map[string]interface{}
					var ov2 map[string]interface{}
					v2, ok = v.(map[string]interface{})
					if ok {
						ov2, ok = osettingValue.(map[string]interface{})
						if ok {
							MergeMap(v2, ov2)
							bplugin.Settings[settingName] = v2
						}
					}
				}
				if !ok {
					bplugin.Settings[settingName] = osettingValue
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
