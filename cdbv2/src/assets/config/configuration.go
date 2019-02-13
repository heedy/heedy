package config

import (
	"fmt"
	"reflect"
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

type ExecJob struct {
	Description *string   `hcl:"description" json:"description,omitempty"`
	Cron        *string   `hcl:"cron" json:"cron,omitempty"`
	Port        *int      `hcl:"port" json:"port,omitempty"`
	KeepAlive   *bool     `hcl:"keepalive" json:"keepalive,omitempty"`
	Cmd         *[]string `hcl:"cmd" json:"cmd,omitempty"`
}

type Plugin struct {
	Version     *string            `hcl:"version" json:"version,omitempty"`
	Description *string            `hcl:"description" json:"description,omitempty"`
	Homepage    *string            `hcl:"homepage" json:"homepage,omitempty"`
	License     *string            `hcl:"license" json:"license,omitempty"`
	GRPC        *string            `hcl:"grpc" json:"grpc,omitempty"`
	Routes      *map[string]string `json:"routes,omitempty"`

	Exec     map[string]*ExecJob `json:"exec,omitempty"`
	Settings map[string]*Setting `json:"settings,omitempty"`
}

func (p *Plugin) Copy() *Plugin {
	np := *p
	np.Exec = make(map[string]*ExecJob)
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

type Group struct {
	GRPC     *bool `json:"grpc,omitempty"`
	REST     *bool `json:"rest,omitempty"`
	Settings *bool `json:"settings,omitempty"`
	AddUser  *bool `json:"add_user,omitempty"`

	// Now, we add a mechanism to "inject" the group into each permissions field
	// This allows simple creation of various levels of admins
	ListGroup *bool `hcl:"list_group" json:"list_group,omitempty"`
	EditGroup *bool `hcl:"edit_group" json:"edit_group,omitempty"`
	DelGroup  *bool `hcl:"del_group" json:"del_group,omitempty"`

	ReadStream  *bool `hcl:"read_stream" json:"read_stream,omitempty"`
	WriteStream *bool `hcl:"write_stream" json:"write_stream,omitempty"`
	ModStream   *bool `hcl:"mod_stream" json:"mod_stream,omitempty"`
	ListStream  *bool `hcl:"list_stream" json:"list_stream,omitempty"`
	EditStream  *bool `hcl:"edit_stream" json:"edit_stream,omitempty"`
	DelStream   *bool `hcl:"del_stream" json:"del_stream,omitempty"`
}

type Configuration struct {
	SiteURL         *string            `hcl:"site_url" json:"site_url,omitempty"`
	Host            *string            `hcl:"host" json:"host,omitempty"`
	Port            *uint16            `hcl:"port" json:"port,omitempty"`
	HTTPPort        *int               `hcl:"http_port" json:"http_port,omitempty"`
	CORS            *bool              `hcl:"cors" json:"cors,omitempty"`
	ActivePlugins   *[]string          `hcl:"plugins" json:"plugins,omitempty"`
	ForbiddenGroups *[]string          `json:"forbidden_groups,omitempty"`
	Plugins         map[string]*Plugin `json:"plugin"`

	Groups map[string]*Group `json:"groups,omitempty"`
}

func (c *Configuration) Copy() *Configuration {
	nc := *c

	nc.Plugins = make(map[string]*Plugin)

	for pkey, pval := range c.Plugins {
		nc.Plugins[pkey] = pval.Copy()
	}

	nc.Groups = make(map[string]*Group)
	for gkey, gval := range c.Groups {
		newg := *gval
		nc.Groups[gkey] = &newg
	}

	return &nc

}

func NewConfiguration() *Configuration {
	return &Configuration{Plugins: make(map[string]*Plugin), Groups: make(map[string]*Group)}
}

func NewPlugin() *Plugin {
	return &Plugin{
		Exec:     make(map[string]*ExecJob),
		Settings: make(map[string]*Setting),
	}
}

// Merges two configurations together
func MergeConfig(base *Configuration, overlay *Configuration) *Configuration {
	base = base.Copy()
	overlay = overlay.Copy()

	CopyStructIfPtrSet(base, overlay)

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

// WriteConfig writes the updates available in the given configuration to the given file.
// It overwrites just the updated values, leaving all others intact
func WriteConfig(filename string, c *Configuration) error {
	/*
		f, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		writer, diag := hclwrite.ParseConfig(f, filename, hcl.Pos{Line: 1, Column: 1})
		if diag != nil {
			return diag
		}
		body := writer.Body()

		// Aaaand we're fucked, because we can't write into blocks
	*/
	return fmt.Errorf("Writer not implemented")
}
