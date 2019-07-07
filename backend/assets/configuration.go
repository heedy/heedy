package assets

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/xeipuuv/gojsonschema"
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
	Version     *string `hcl:"version" json:"version,omitempty"`
	Description *string `hcl:"description" json:"description,omitempty"`
	Homepage    *string `hcl:"homepage" json:"homepage,omitempty"`
	License     *string `hcl:"license" json:"license,omitempty"`

	Frontend *string            `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`
	Backend  *map[string]string `json:"backend,omitempty"`

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

type SourceType struct {
	Frontend *string            `json:"frontend,omitempty" hcl:"frontend,block" cty:"frontend"`
	Backend  *map[string]string `json:"backend,omitempty" hcl:"backend" cty:"backend"`

	Meta *map[string]interface{} `json:"meta,omitempty"`

	Scopes *map[string]string `json:"scopes,omitempty" hcl:"scopes" cty:"scopes"`

	metaSchema *gojsonschema.Schema
	metaObj    map[string]interface{}
}

func (s *SourceType) Copy() SourceType {
	snew := SourceType{}
	CopyStructIfPtrSet(&snew, s)
	if s.Backend != nil {
		newRoutes := make(map[string]string)
		for k, v := range *(s.Backend) {
			newRoutes[k] = v
		}
		snew.Backend = &newRoutes
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

	for k, v := range c.SourceTypes {
		err := v.ValidateMeta(nil)
		if err != nil {
			return fmt.Errorf("source %s meta schema invalid: %s", k, err.Error())
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
			// Copy the scopes to av
			if av.Scopes != nil && cv.Scopes != nil {
				for sk, sv := range *av.Scopes {
					(*cv.Scopes)[sk] = sv
				}
				av.Scopes = cv.Scopes
			}
			// Copy the routes to av
			if av.Backend != nil && cv.Backend != nil {
				for rk, rv := range *av.Backend {
					(*cv.Backend)[rk] = rv
				}
				av.Backend = cv.Backend
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
