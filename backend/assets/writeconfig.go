package assets

import (
	"io/ioutil"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// WriteConfig writes the updates available in the given configuration to the given file.
// It overwrites just the updated values, leaving all others intact
func WriteConfig(filename string, c *Configuration) error {

	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	writer, diag := hclwrite.ParseConfig(f, filename, hcl.Pos{Line: 1, Column: 1})
	if diag != nil {
		return diag
	}
	body := writer.Body()

	if c.URL != nil {
		if strings.HasSuffix(*c.URL, "/") {
			noslash := (*c.URL)[:len(*c.URL)-1]
			c.URL = &noslash
		}
		body.SetAttributeValue("url", cty.StringVal(*c.URL))
	}
	if c.Host != nil {
		body.SetAttributeValue("host", cty.StringVal(*c.Host))
	}
	if c.Port != nil {
		body.SetAttributeValue("port", cty.NumberIntVal(int64(*c.Port)))
	}
	if c.ActivePlugins != nil {
		plist := make([]cty.Value, 0)
		for i := range *c.ActivePlugins {
			plist = append(plist, cty.StringVal((*c.ActivePlugins)[i]))
		}
		if len(plist) > 0 {
			body.SetAttributeValue("plugins", cty.ListVal(plist))
		} else {
			body.SetAttributeValue("plugins", cty.ListValEmpty(cty.String))
		}
	}
	if c.AdminUsers != nil {
		alist := make([]cty.Value, 0)
		for i := range *c.AdminUsers {
			alist = append(alist, cty.StringVal((*c.AdminUsers)[i]))
		}
		if len(alist) > 0 {
			body.SetAttributeValue("admin_users", cty.ListVal(alist))
		} else {
			body.SetAttributeValue("admin_users", cty.ListValEmpty(cty.String))
		}

	}

	for pname, p := range c.Plugins {
		blk := body.FirstMatchingBlock("plugin", []string{pname})
		if blk == nil {
			blk = body.AppendNewBlock("plugin", []string{pname})
		}
		b := blk.Body()
		for sname, svalue := range p.Settings {
			var v cty.Value
			switch sv := svalue.(type) {
			case int:
				v = cty.NumberIntVal(int64(sv))
			case int64:
				v = cty.NumberIntVal(sv)
			case float64:
				v = cty.NumberFloatVal(sv)
			case string:
				v = cty.StringVal(sv)
			case []string:
				if len(sv) == 0 {
					v = cty.ListValEmpty(cty.String)
				} else {
					slist := make([]cty.Value, 0, len(sv))
					for i := range sv {
						slist = append(slist, cty.StringVal(sv[i]))
					}
					v = cty.ListVal(slist)
				}
			}
			b.SetAttributeValue(sname, v)
		}
	}

	return ioutil.WriteFile(filename, writer.Bytes(), 0755)
}
