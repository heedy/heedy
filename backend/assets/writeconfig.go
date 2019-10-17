package assets

import (
	"io/ioutil"

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

	// Aaaand we're fucked, because we can't write into blocks
	if c.SiteURL != nil {
		body.SetAttributeValue("site_url", cty.StringVal(*c.SiteURL))
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

	return ioutil.WriteFile(filename, writer.Bytes(), 0755)
}
