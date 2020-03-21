package assets

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"path"
	"path/filepath"
	"runtime"

	"github.com/heedy/heedy/backend/buildinfo"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type hclRun struct {
	Name string  `hcl:"name,label"`
	Type *string `hcl:"type"`

	Enabled *bool   `hcl:"enabled" json:"enabled,omitempty"`
	Cron    *string `hcl:"cron" json:"cron,omitempty"`

	// Everything that remains is settings specific to the runner
	Settings hcl.Body `hcl:",remain"`
}

type hclObject struct {
	Key  string `hcl:"key,label"`
	Name string `hcl:"name"`
	Type string `hcl:"type"`

	AutoCreate *bool `json:"auto_create,omitempty" hcl:"auto_create"`

	Description *string `hcl:"description"`
	Icon        *string `hcl:"icon"`
	OwnerScope  *string `hcl:"owner_scope"`
	Tags        *string `hcl:"tags"`

	Meta *cty.Value `hcl:"meta,attr"`

	On []Event `hcl:"on,block" json:"on,omitempty"`
}

type hclApp struct {
	Plugin string `hcl:"plugin,label"`
	Name   string `hcl:"name"`

	AutoCreate  *bool `json:"auto_create,omitempty" hcl:"auto_create"`
	Unique      *bool `json:"unique,omitempty" hcl:"unique"`
	AccessToken *bool `json:"access_token,omitempty" hcl:"access_token"`

	Description *string   `json:"description" hcl:"description"`
	Icon        *string   `json:"icon" hcl:"icon"`
	Scope       *string   `json:"scope,omitempty" hcl:"scope"`
	Enabled     *bool     `json:"enabled,omitempty" hcl:"enabled"`
	Readonly    *[]string `json:"readonly,omitempty" hcl:"readonly"`

	Settings       *cty.Value `hcl:"settings,attr"`
	SettingsSchema *cty.Value `hcl:"settings_schema,attr"`

	Objects []hclObject `hcl:"object,block"`
	On      []Event     `hcl:"on,block" json:"on,omitempty"`
}

type hclPlugin struct {
	Name        string  `hcl:"name,label"`
	Icon        *string `hcl:"icon" json:"icon"`
	Version     *string `hcl:"version" json:"version"`
	Description *string `hcl:"description" json:"description"`
	Homepage    *string `hcl:"homepage" json:"homepage"`
	License     *string `hcl:"license" json:"license"`

	Frontend *string `hcl:"frontend" json:"frontend"`

	Routes *map[string]string `hcl:"routes" json:"routes"`
	Events *map[string]string `hcl:"events" json:"events,omitempty"`

	SettingSchema *cty.Value `hcl:"settings_schema"`

	Run []hclRun `hcl:"run,block"`

	Apps []hclApp `hcl:"app,block"`
	On   []Event  `hcl:"on,block" json:"on,omitempty"`

	// The remaining stuff is plugin-specific settings
	// that will be passed to the plugin executables,
	// and can be queried by javascript as part of the configuration
	Settings hcl.Body `hcl:",remain"`
}

type hclObjectType struct {
	Label string `hcl:"label,label"`

	Frontend *string `json:"frontend,omitempty" hcl:"frontend" cty:"frontend"`

	Routes *map[string]string `json:"routes,omitempty" hcl:"routes" cty:"routes"`

	Meta  *cty.Value         `hcl:"meta,attr"`
	Scope *map[string]string `json:"scope,omitempty" hcl:"scope" cty:"scope"`
}

type hclRunType struct {
	Label  string     `hcl:"label,label"`
	Schema *cty.Value `hcl:"schema,attr"`
	API    *string    `json:"api,omitempty" hcl:"api" cty:"api"`
}

type hclConfiguration struct {
	URL            *string   `hcl:"url" json:"url,omitempty"`
	Host           *string   `hcl:"host" json:"host,omitempty"`
	Port           *uint16   `hcl:"port" json:"port,omitempty"`
	ActivePlugins  *[]string `hcl:"active_plugins" json:"active_plugins,omitempty"`
	AdminUsers     *[]string `hcl:"admin_users" json:"admin_users,omitempty"`
	ForbiddenUsers *[]string `hcl:"forbidden_users" json:"forbidden_users,omitempty"`

	Language         *string `hcl:"language" json:"language"`
	FallbackLanguage *string `hcl:"fallback_language" json:"fallback_language"`

	SQL *string `hcl:"sql" json:"sql,omitempty"`

	Frontend *string `hcl:"frontend"`

	RunTimeout *string `hcl:"run_timeout"`

	Scope       *map[string]string `json:"scope,omitempty" hcl:"scope"`
	NewAppScope *[]string          `json:"new_app_scope,omitempty" hcl:"new_app_scope"`

	ObjectTypes []hclObjectType `json:"type" hcl:"type,block"`
	RunTypes    []hclRunType    `json:"runtype" hcl:"runtype,block"`

	RequestBodyByteLimit *int64 `hcl:"request_body_byte_limit" json:"request_body_byte_limit,omitempty"`
	AllowPublicWebsocket *bool  `hcl:"allow_public_websocket" json:"allow_public_websocket,omitempty"`

	Plugins []hclPlugin `hcl:"plugin,block"`

	LogLevel *string `json:"log_level" hcl:"log_level"`
	LogFile  *string `json:"log_file" hcl:"log_file"`
}

func loadJSONObject(v *cty.Value) (*map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}
	var obj map[string]interface{}
	b, err := json.Marshal(ctyjson.SimpleJSONValue{Value: *v})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &obj)
	return &obj, err
}

func loadJSONObjectBody(bdy hcl.Body, ctx *hcl.EvalContext) (map[string]interface{}, error) {
	// And now, finally, read in the setting values for the plugin
	settings := make(map[string]*hcl.Attribute)

	// Now, there will be an error talking about "no exec block allowed blah blah"
	// This is BS, and we don't care.
	gohcl.DecodeBody(bdy, ctx, &settings)

	jsonSettings := make(map[string]ctyjson.SimpleJSONValue)
	for key, attr := range settings {
		val, err := attr.Expr.Value(ctx)
		if err != nil {
			return nil, err
		}
		jsonSettings[key] = ctyjson.SimpleJSONValue{Value: val}

	}

	var obj map[string]interface{}
	b, err := json.Marshal(jsonSettings)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &obj)
	return obj, err
}

func loadConfigFromHcl(f *hcl.File, filename string) (*Configuration, error) {
	// The configuration is initially unmarshalled into the hclConfiguration
	// object, which then needs extra processing to get into the format that heedy
	// can use.
	// TODO: Perhaps it might be helpful to fix these issues uptimeseries,
	// 			rather than messing with a bunch of workarounds?
	hc := &hclConfiguration{}

	// We set up the parsing context with the useful variables and functions
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"os":      cty.StringVal(runtime.GOOS),
			"arch":    cty.StringVal(runtime.GOARCH),
			"version": cty.StringVal(buildinfo.Version),
		},
		Functions: map[string]function.Function{
			"jsondecode": stdlib.JSONDecodeFunc,
			"jsonencode": stdlib.JSONEncodeFunc,
			"concat":     stdlib.ConcatFunc,
			"format":     stdlib.FormatFunc,
			"int":        stdlib.IntFunc,
			"join": function.New(&function.Spec{
				VarParam: &function.Parameter{
					Name: "string",
					Type: cty.String,
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					result := ""
					for _, s := range args {
						result += s.AsString()
					}
					return cty.StringVal(string(result)), nil
				},
			}),
			"file": function.New(&function.Spec{
				Params: []function.Parameter{
					{
						Name: "filename",
						Type: cty.String,
					},
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					readFile := args[0].AsString()
					readFile = path.Join(path.Dir(filename), readFile)
					b, err := ioutil.ReadFile(readFile)
					if err != nil {
						return cty.StringVal(""), err
					}
					return cty.StringVal(string(b)), nil
				},
			}),
			"datauri": function.New(&function.Spec{
				Params: []function.Parameter{
					{
						Name: "filename",
						Type: cty.String,
					},
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					readFile := args[0].AsString()
					readFile = path.Join(path.Dir(filename), readFile)
					b, err := ioutil.ReadFile(readFile)
					if err != nil {
						return cty.StringVal(""), err
					}

					datauri := fmt.Sprintf("data:%s;base64,%s", mime.TypeByExtension(filepath.Ext(readFile)), base64.StdEncoding.EncodeToString(b))

					return cty.StringVal(datauri), nil
				},
			}),
		},
	}

	diag := gohcl.DecodeBody(f.Body, ctx, hc)
	if diag != nil {
		return nil, diag
	}

	// Now we move the values over to the configuration
	c := NewConfiguration()
	CopyStructIfPtrSet(c, hc)

	// Loop through the objects
	for _, ht := range hc.ObjectTypes {
		t := ObjectType{}

		CopyStructIfPtrSet(&t, ht)

		var err error
		t.Meta, err = loadJSONObject(ht.Meta)
		if err != nil {
			return nil, err
		}

		c.ObjectTypes[ht.Label] = t
	}

	for _, v := range hc.RunTypes {
		r := RunType{}
		CopyStructIfPtrSet(&r, &v)

		if v.Schema != nil {
			m, err := loadJSONObject(v.Schema)
			if err != nil {
				return nil, err
			}
			r.Schema = *m
		} else {
			r.Schema = make(map[string]interface{})
		}

		c.RunTypes[v.Label] = r
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

		for j := range hp.Run {
			if hp.Run[j].Name == "" {
				return nil, fmt.Errorf("%s: Plugin %s no label on run", filename, hp.Name)
			}
			if _, ok := p.Run[hp.Run[j].Name]; ok {
				return nil, fmt.Errorf("%s: Plugin %s run %s defined twice", filename, hp.Name, hp.Run[j].Name)
			}

			ej := Run{}

			CopyStructIfPtrSet(&ej, &hp.Run[j])

			o, err := loadJSONObjectBody(hp.Run[j].Settings, ctx)
			if err != nil {
				return nil, err
			}
			ej.Settings = o

			p.Run[hp.Run[j].Name] = ej
		}
		for o := range hp.On {
			if err := hp.On[o].Validate(); err != nil {
				return nil, fmt.Errorf("%s: Plugin %s - %w", filename, hp.Name, err)
			}
			if hp.On[o].Event == "" {
				return nil, fmt.Errorf("%s: Plugin %s 'on' without event", filename, hp.Name)
			}
		}
		p.On = hp.On
		if hp.SettingSchema != nil {
			sobj, err := loadJSONObject(hp.SettingSchema)
			if err != nil {
				return nil, err
			}
			p.SettingsSchema = *sobj
		}

		// Load the apps that the plugin wants to set up
		for j := range hp.Apps {
			hc := hp.Apps[j]
			if _, ok := p.Apps[hc.Plugin]; ok {
				return nil, fmt.Errorf("%s: Plugin %s app %s defined twice", filename, hp.Name, hc.Plugin)
			}
			conn := NewApp()
			conn.Name = hp.Apps[j].Name
			CopyStructIfPtrSet(conn, &hc)
			var err error
			conn.Settings, err = loadJSONObject(hc.Settings)
			if err != nil {
				return nil, err
			}
			conn.SettingsSchema, err = loadJSONObject(hc.SettingsSchema)
			if err != nil {
				return nil, err
			}
			for oi := range hc.On {
				o := hc.On[oi]
				if err := o.Validate(); err != nil {
					return nil, fmt.Errorf("%s: Plugin %s - %w", filename, hp.Name, err)
				}
				if o.Event == "" {
					return nil, fmt.Errorf("%s: Plugin %s app %s 'on' without event", filename, hp.Name, conn.Name)
				}
			}
			conn.On = hc.On
			for k := range hc.Objects {
				hs := hc.Objects[k]
				if hs.Key == "" {
					return nil, fmt.Errorf("%s: Plugin %s app %s object with no label", filename, hp.Name, hc.Plugin)
				}
				if _, ok := conn.Objects[hs.Key]; ok {
					return nil, fmt.Errorf("%s: Plugin %s app %s object %s defined twice", filename, hp.Name, hc.Plugin, hs.Key)
				}
				s := NewObject()
				s.Name = hs.Name
				s.Type = hs.Type

				CopyStructIfPtrSet(s, &hs)
				s.Meta, err = loadJSONObject(hs.Meta)
				if err != nil {
					return nil, err
				}
				for oi := range hs.On {
					o := hs.On[oi]
					if err := o.Validate(); err != nil {
						return nil, fmt.Errorf("%s: Plugin %s - %w", filename, hp.Name, err)
					}
					if o.Event == "" {
						return nil, fmt.Errorf("%s: Plugin %s app %s object %s 'on' without event", filename, hp.Name, conn.Name, s.Name)
					}
				}
				s.On = hs.On

				conn.Objects[hs.Key] = s
			}

			p.Apps[hc.Plugin] = conn
		}

		obj, err := loadJSONObjectBody(hp.Settings, ctx)
		if err != nil {
			return nil, err
		}
		p.Settings = obj

		c.Plugins[hp.Name] = p

	}

	//b, _ := json.MarshalIndent(c, "", "  ")
	//fmt.Printf("\n%s\n\n", string(b))

	return c, nil
}
