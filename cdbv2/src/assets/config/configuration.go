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
	Description *string `hcl:"description" json:"description,omitempty"`
	Cron        *string `hcl:"cron" json:"cron,omitempty"`
	Cmd         *string `hcl:"cmd" json:"cmd,omitempty"`
}

type Plugin struct {
	Cmd         *string `hcl:"cmd" json:"cmd,omitempty"`
	Version     *string `hcl:"version" json:"version,omitempty"`
	Description *string `hcl:"description" json:"description,omitempty"`
	Homepage    *string `hcl:"homepage" json:"homepage,omitempty"`
	License     *string `hcl:"license" json:"license,omitempty"`

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

type Configuration struct {
	SiteURL       *string            `hcl:"siteurl" json:"siteurl,omitempty"`
	Port          *uint16            `hcl:"port" json:"port,omitempty"`
	ActivePlugins *[]string          `hcl:"plugins" json:"plugins,omitempty"`
	Plugins       map[string]*Plugin `json:"plugin"`
}

func (c *Configuration) Copy() *Configuration {
	nc := *c

	nc.Plugins = make(map[string]*Plugin)

	for pkey, pval := range c.Plugins {
		nc.Plugins[pkey] = pval.Copy()
	}

	return &nc

}

func NewConfiguration() *Configuration {
	return &Configuration{Plugins: make(map[string]*Plugin)}
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
