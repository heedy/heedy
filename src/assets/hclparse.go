package assets

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

type hclExecJob struct {
	Name string `hcl:"name,label"`

	Description *string   `hcl:"description" json:"description,omitempty"`
	Cron        *string   `hcl:"cron" json: "cron,omitempty"`
	Port        *int      `hcl:"port"`
	KeepAlive   *bool     `hcl:"keepalive"`
	Cmd         *[]string `hcl:"cmd" json: "cmd,omitempty"`
}

type hclPlugin struct {
	Name           string                    `hcl:"name,label"`
	Version        *string                   `hcl:"version" json:"version"`
	Description    *string                   `hcl:"description" json:"description"`
	Homepage       *string                   `hcl:"homepage" json:"homepage"`
	License        *string                   `hcl:"license" json:"license"`
	GRPC           *string                   `hcl:"grpc" json:"grpc"`
	Routes         *map[string]string        `hcl:"routes" json:"routes"`
	SettingSchemas *map[string]hclJSONSchema `hcl:"settings"`
	//FallbackLanguage *string                   `hcl:"fallback_language" json:"fallback_language"`

	Exec []hclExecJob `hcl:"exec,block"`

	// The remaining stuff is plugin-specific settings
	// that will be passed to the plugin executables,
	// and can be queried by javascript as part of the configuration
	Settings hcl.Body `hcl:",remain"`
}

type hclGroup struct {
	Name     string `hcl:"name,label"`
	GRPC     *bool  `hcl:"grpc" json:"grpc,omitempty"`
	REST     *bool  `hcl:"rest"`
	Settings *bool  `hcl:"settings" json:"settings,omitempty"`
	AddUser  *bool  `hcl:"add_user" json:"add_user,omitempty"`

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

type hclConfiguration struct {
	SiteURL       *string   `hcl:"site_url" json:"site_url,omitempty"`
	Host          *string   `hcl:"host" json:"host,omitempty"`
	Port          *uint16   `hcl:"port" json:"port,omitempty"`
	HTTPPort      *uint16   `hcl:"http_port" json:"http_port,omitempty"`
	CORS          *bool     `hcl:"cors"`
	ActivePlugins *[]string `hcl:"plugins"`

	SQL *string `hcl:"sql" json:"sql,omitempty"`

	Language         *string `hcl:"language" json:"language"`
	FallbackLanguage *string `hcl:"fallback_language" json:"fallback_language"`

	Plugins []hclPlugin `hcl:"plugin,block"`
	Groups  []hclGroup  `hcl:"group,block"`

	RequestBodyByteLimit *int64 `hcl:"request_body_byte_limit" json:"request_body_byte_limit,omitempty"`
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
	// object, which then needs extra processing to get into the format that ConnectorDB
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
	/*
		c.SiteURL = hc.SiteURL
		c.Port = hc.Port
		c.ActivePlugins = hc.ActivePlugins
	*/

	// Loop through the groups
	for i := range hc.Groups {
		hg := hc.Groups[i]
		if hg.Name == "" {
			return nil, fmt.Errorf("%s: Can't use group with no name", filename)
		}
		if _, ok := c.Groups[hg.Name]; ok {
			return nil, fmt.Errorf("%s: Group \"%s\" defined twice", filename, hg.Name)
		}

		g := &Group{}

		CopyStructIfPtrSet(g, hg)

		c.Groups[hg.Name] = g
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

			ej := &ExecJob{}
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

			currentSetting.Value = val.AsString()

			p.Settings[key] = currentSetting
		}

		c.Plugins[hp.Name] = p

	}

	return c, nil
}
