package assets

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

var (
	configparser = hclparse.NewParser()
)

type hclJSONSchema struct {
	Title       *string `hcl:"title" json:"title,omitempty" cty:"title"`
	Type        *string `hcl:"type" json:"type,omitempty" cty:"type"`
	Description *string `hcl:"description" json:"description,omitempty" cty:"description"`
	//Minimum          *float64 `hcl:"minimum" json:"minimum,omitempty" cty:"minimum"`
	//ExclusiveMinimum *float64 `hcl:"exclusiveMinimum" json:"exclusiveMinimum,omitempty" cty:"exclusiveMinimum"`
	//Maximum          *float64 `hcl:"maximum" json:"maximum,omitempty" cty:"maximum"`
	//ExclusiveMaximum *float64 `hcl:"exclusiveMaximum" json:"exclusiveMaximum,omitempty" cty:"exclusiveMaximum"`
	//Items            *JSONSchema    `hcl:"items" json:"items,omitempty"`
	//MinItems         *int           `hcl:"minItems" json:"minItems,omitempty"`
	//UniqueItems      *bool          `hcl:"uniqueItems" json:"uniqueItems,omitempty"`
	//Default *hcl.Attribute `hcl:"default"`
}

type hclExec struct {
	Name string `hcl:"name,label"`

	Enabled   *bool     `hcl:"enabled" json:"enabled,omitempty"`
	Cron      *string   `hcl:"cron" json: "cron,omitempty"`
	KeepAlive *bool     `hcl:"keepalive"`
	Cmd       *[]string `hcl:"cmd" json: "cmd,omitempty"`
}

type hclPlugin struct {
	Name        string  `hcl:"name,label"`
	Version     *string `hcl:"version" json:"version"`
	Description *string `hcl:"description" json:"description"`
	Homepage    *string `hcl:"homepage" json:"homepage"`
	License     *string `hcl:"license" json:"license"`

	Frontend *string            `hcl:"frontend" json:"frontend"`

	Routes  *map[string]string `hcl:"routes" json:"routes"`

	SettingSchemas *map[string]hclJSONSchema `hcl:"settings"`

	Exec []hclExec `hcl:"exec,block"`

	// The remaining stuff is plugin-specific settings
	// that will be passed to the plugin executables,
	// and can be queried by javascript as part of the configuration
	Settings hcl.Body `hcl:",remain"`
}

type hclSourceType struct {
	Label string `hcl:"label,label"`

	Frontend *string `json:"frontend,omitempty" hcl:"frontend" cty:"frontend"`

	Routes *map[string]string `json:"routes,omitempty" hcl:"routes" cty:"routes"`

	Meta   *cty.Value         `hcl:"meta,attr"`
	Scopes *map[string]string `json:"scopes,omitempty" hcl:"scopes" cty:"scopes"`
}

type hclConfiguration struct {
	SiteURL        *string   `hcl:"site_url" json:"site_url,omitempty"`
	Host           *string   `hcl:"host" json:"host,omitempty"`
	Port           *uint16   `hcl:"port" json:"port,omitempty"`
	ActivePlugins  *[]string `hcl:"plugins" json:"plugins,omitempty"`
	AdminUsers     *[]string `hcl:"admin_users" json:"admin_users,omitempty"`
	ForbiddenUsers *[]string `hcl:"forbidden_users" json:"forbidden_users,omitempty"`

	Language         *string `hcl:"language" json:"language"`
	FallbackLanguage *string `hcl:"fallback_language" json:"fallback_language"`

	SQL *string `hcl:"sql" json:"sql,omitempty"`

	Frontend *string `hcl:"frontend"`

	ExecTimeout *string `hcl:"exec_timeout"`

	Scopes              *map[string]string `json:"scopes,omitempty" hcl:"scopes"`
	NewConnectionScopes *[]string          `json:"new_connection_scopes,omitempty" hcl:"new_connection_scopes"`

	SourceTypes []hclSourceType `json:"source_types" hcl:"source,block"`

	RequestBodyByteLimit *int64 `hcl:"request_body_byte_limit" json:"request_body_byte_limit,omitempty"`

	Plugins []hclPlugin `hcl:"plugin,block"`
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
			//fmt.Println(fieldName)

			baseFieldValue := bv.FieldByName(fieldName)
			if baseFieldValue.IsValid() && baseFieldValue.Type() == fieldValue.Type() {
				if !fieldValue.IsNil() {
					//fmt.Printf("Setting %s\n", fieldName)
					baseFieldValue.Set(fieldValue)
				}

			}

		}
	}

}

func loadConfigFromHcl(f *hcl.File, filename string) (*Configuration, error) {
	// The configuration is initially unmarshalled into the hclConfiguration
	// object, which then needs extra processing to get into the format that heedy
	// can use.
	// TODO: Perhaps it might be helpful to fix these issues upstream,
	// 			rather than messing with a bunch of workarounds?
	hc := &hclConfiguration{}

	diag := gohcl.DecodeBody(f.Body, nil, hc)
	if diag != nil {
		return nil, diag
	}

	// Now we move the values over to the configuration
	c := NewConfiguration()
	CopyStructIfPtrSet(c, hc)

	// Loop through the sources
	for _, ht := range hc.SourceTypes {
		t := SourceType{}

		CopyStructIfPtrSet(&t, ht)

		if ht.Meta != nil {
			var metaMap map[string]interface{}
			b, err := json.Marshal(ctyjson.SimpleJSONValue{Value: *ht.Meta})
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(b, &metaMap)
			if err != nil {
				return nil, err
			}
			t.Meta = &metaMap
		}

		c.SourceTypes[ht.Label] = t
	}

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

			ej := &Exec{}
			CopyStructIfPtrSet(ej, &hp.Exec[j])
			p.Exec[hp.Exec[j].Name] = ej
		}
		if hp.SettingSchemas != nil {
			for k, v := range *hp.SettingSchemas {
				setting := &Setting{}
				//fmt.Println(k,v)
				CopyStructIfPtrSet(setting, &v)
				p.Settings[k] = setting
			}
		}
		/*
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
		*/

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

			if err := gocty.FromCtyValue(val,&currentSetting.Value); err!=nil {
				return nil, fmt.Errorf("%s: Plugin %s attribute '%s' interpreted as custom string setting value, but got error: %w", filename, hp.Name, key, err)
			}

			p.Settings[key] = currentSetting
		}

		c.Plugins[hp.Name] = p

	}

	return c, nil
}
