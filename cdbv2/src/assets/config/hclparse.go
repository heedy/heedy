package config

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
)

var (
	configparser = hclparse.NewParser()
)

type hclJSONSchema struct {
	Name string `hcl:"name,label"`

	Title            *string  `hcl:"title" json:"title,omitempty"`
	Type             *string  `hcl:"type" json:"type,omitempty"`
	Description      *string  `hcl:"description" json:"description,omitempty"`
	Minimum          *float64 `hcl:"minimum" json:"minimum,omitempty"`
	ExclusiveMinimum *float64 `hcl:"exclusiveMinimum" json:"exclusiveMinimum,omitempty"`
	Maximum          *float64 `hcl:"maximum" json:"maximum,omitempty"`
	ExclusiveMaximum *float64 `hcl:"exclusiveMaximum" json:"exclusiveMaximum,omitempty"`
	//Items            *JSONSchema    `hcl:"items" json:"items,omitempty"`
	//MinItems         *int           `hcl:"minItems" json:"minItems,omitempty"`
	//UniqueItems      *bool          `hcl:"uniqueItems" json:"uniqueItems,omitempty"`
	Default *hcl.Attribute `hcl:"default"`
}

type hclExecJob struct {
	Name string `hcl:"name,label"`

	Description *string `hcl:"description" json:"description,omitempty"`
	Cron        *string `hcl:"cron" json: "cron,omitempty"`
	Cmd         *string `hcl:"cmd" json: "cmd,omitempty"`
}

type hclPlugin struct {
	Name        string  `hcl:"name,label"`
	Cmd         *string `hcl:"cmd" json:"cmd"`
	Version     *string `hcl:"version" json:"version"`
	Description *string `hcl:"description" json:"description"`
	Homepage    *string `hcl:"homepage" json:"homepage"`
	License     *string `hcl:"license" json:"license"`

	Exec []hclExecJob `hcl:"exec,block"`

	SettingSchema []hclJSONSchema `hcl:"setting,block"`

	// The remaining stuff is plugin-specific settings
	// that will be passed to the plugin executables,
	// and can be queried by javascript as part of the configuration
	Settings hcl.Body `hcl:",remain"`
}

type hclConfiguration struct {
	SiteURL       *string     `hcl:"siteurl"`
	Port          *uint16     `hcl:"port"`
	ActivePlugins *[]string   `hcl:"plugins"`
	Plugins       []hclPlugin `hcl:"plugin,block"`
}

func preprocess(i interface{}) (reflect.Value, reflect.Kind) {
	v := reflect.ValueOf(i)
	k := v.Kind()
	for k == reflect.Ptr {
		v = reflect.Indirect(v)
		k = v.Kind()
	}
	return v, k
}

// CopyIfSet copies all pointer params from overlay to base
// Does not touch arrays and things that don't have identical types
func CopyStructIfPtrSet(base interface{}, overlay interface{}) {
	bv, _ := preprocess(base)
	ov, _ := preprocess(overlay)

	tot := ov.NumField()
	for i := 0; i < tot; i++ {
		// Now check if the field is of type ptr
		fieldValue := ov.Field(i)

		if fieldValue.Kind() == reflect.Ptr {
			// Only if it is a ptr do we continue, since that's all that we care about
			fieldName := ov.Type().Field(i).Name

			baseFieldValue := bv.FieldByName(fieldName)
			if baseFieldValue.Type() == fieldValue.Type() {
				if !fieldValue.IsNil() {
					fmt.Printf("Setting %s\n", fieldName)
					baseFieldValue.Set(fieldValue)
				}

			}

		}
	}

}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Configuration, error) {

	f, diag := configparser.ParseHCLFile(filename)
	if diag != nil {
		return nil, diag
	}

	// The configuration is initially unmarshalled into the hclConfiguration
	// object, which then needs extra processing to get into the format that ConnectorDB
	// can use.
	// TODO: Perhaps it might be helpful to fix these issues upstream,
	// 			rather than messing with a bunch of workarounds?
	hc := &hclConfiguration{}

	diag = gohcl.DecodeBody(f.Body, nil, hc)
	if diag != nil {
		return nil, diag
	}

	// Now we move the values over to the configuration
	c := NewConfiguration()
	CopyStructIfPtrSet(c, hc)
	/*
		c.SiteURL = hc.SiteURL
		c.Port = hc.Port
		c.ActivePlugins = hc.ActivePlugins
	*/

	// Loop through the plugins
	for i := range hc.Plugins {
		hp := hc.Plugins[i]
		if hp.Name == "" {
			return nil, fmt.Errorf("%s: Can't use plugin with no name", filename)
		}

		if _, ok := c.Plugins[hp.Name]; ok {
			return nil, fmt.Errorf("%s: Plugin \"%s\" defined twice", filename, hp.Name)
		}

		p := NewPlugin()

		CopyStructIfPtrSet(p, hp)

		for j := range hp.Exec {
			if hp.Exec[j].Name == "" {
				return nil, fmt.Errorf("%s: Plugin %s no label on exec", filename, hp.Name)
			}
			if _, ok := p.Exec[hp.Exec[j].Name]; ok {
				return nil, fmt.Errorf("%s: Plugin %s exec %s defined twice", filename, hp.Name, hp.Exec[j].Name)
			}

			ej := &ExecJob{}
			CopyStructIfPtrSet(ej, &hp.Exec[j])
			p.Exec[hp.Exec[j].Name] = ej
		}

		for j := range hp.SettingSchema {
			if hp.SettingSchema[j].Name == "" {
				return nil, fmt.Errorf("%s: Plugin %s has missing label on setting", filename, hp.Name)
			}

			hj := hp.SettingSchema[j]

			setting := &Setting{}
			CopyStructIfPtrSet(setting, &hj)
			if hj.Default != nil {
				// There is an attribute there, so read it into a string
				val, diag := hj.Default.Expr.Value(nil)
				if diag != nil {
					return nil, diag
				}
				setting.Default = val.AsString()
			}

			p.Settings[hp.SettingSchema[j].Name] = setting
		}

		// And now, finally, read in the setting values
		settings := make(map[string]*hcl.Attribute)

		// Now, there will be an error talking about "no exec block allowed blah blah"
		// This is BS, and we don't care.
		gohcl.DecodeBody(hp.Settings, nil, &settings)

		for key, attr := range settings {

			// Now we read in the actual settings. If the key does not exist, create one for this setting
			currentSetting, ok := p.Settings[key]
			if !ok {
				// There is no such setting defined. Work around it by defining one
				currentSetting = &Setting{}
			}

			val, diag := attr.Expr.Value(nil)
			if diag != nil {
				return nil, diag
			}

			currentSetting.Value = val.AsString()

			p.Settings[key] = currentSetting
		}

		c.Plugins[hp.Name] = p

	}

	return c, nil
}
