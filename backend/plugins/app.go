package plugins

import (
	"errors"
	"strings"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/database/dbutil"
)

func App(pluginKey string, owner string, cv *assets.App) *database.App {
	c := &database.App{
		Details: database.Details{
			Name:        &cv.Name,
			Description: cv.Description,
			Icon:        cv.Icon,
		},
		Enabled: cv.Enabled,
		Plugin:  &pluginKey,
		Owner:   &owner,
	}
	if cv.Scope != nil {
		c.Scope = &database.AppScopeArray{
			ScopeArray: database.ScopeArray{},
		}
		c.Scope.Load(*cv.Scope)
	}
	if cv.AccessToken == nil || !(*cv.AccessToken) {
		empty := ""
		c.AccessToken = &empty
	}
	if cv.SettingsSchema != nil {
		jo := dbutil.JSONObject(*cv.SettingsSchema)
		c.SettingsSchema = &jo
	}
	if cv.Settings != nil {
		jo := dbutil.JSONObject(*cv.Settings)
		c.Settings = &jo
	}
	return c
}

func AppObject(app string, key string, as *assets.Object) *database.Object {
	s := &database.Object{
		Details: database.Details{
			Name:        &as.Name,
			Description: as.Description,
			Icon:        as.Icon,
		},
		App:  &app,
		Key:  &key,
		Type: &as.Type,
	}
	if as.Tags != nil {
		s.Tags = &dbutil.StringArray{}
		s.Tags.Load(*as.Tags)
	}
	if as.Meta != nil {
		jo := dbutil.JSONObject(*as.Meta)
		s.Meta = &jo
	}
	if as.OwnerScope != nil {
		s.OwnerScope = &database.ScopeArray{}
		s.OwnerScope.Load(*as.OwnerScope)
	}

	return s
}

func CreateApp(c *rest.Context, app *database.App) (string, string, error) {
	if c.DB.Type() != database.UserType && c.DB.Type() != database.AdminType {
		return "", "", database.ErrAccessDenied("Only users can create apps")
	}
	owner := ""
	if app.Owner != nil {
		owner = *app.Owner
	}
	if c.DB.Type() == database.UserType && owner == "" {
		owner = c.DB.ID()
	}
	if c.DB.Type() == database.UserType && owner != c.DB.ID() {
		return "", "", database.ErrAccessDenied("You can only create an app for your own user")
	}
	if owner == "" {
		return "", "", errors.New("App must have an owner")
	}
	pk := strings.Split(*app.Plugin, ":")
	if len(pk) != 2 {
		return "", "", database.ErrBadQuery("plugin keys must be in the format 'plugin_name:key'")
	}
	adb := c.DB.AdminDB()
	a := adb.Assets()

	p, ok := a.Config.Plugins[pk[0]]
	if !ok {
		return "", "", database.ErrBadQuery("unrecognized plugin name for app plugin")
	}

	papp, ok := p.Apps[pk[1]]
	if !ok {
		// The plugin doesn't have this app - if trying to create as a user, fail,
		// if as an admin... just create it.
		if c.DB.Type() == database.AdminType {
			return adb.CreateApp(app)
		}
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}

	// Check if this key is from an *active* plugin

	ap := a.Config.GetActivePlugins()
	hadPlugin := false
	for _, p := range ap {
		if p == pk[0] {
			hadPlugin = true
			break
		}
	}
	if !hadPlugin {
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}

	if papp.Unique != nil && *papp.Unique {
		a, err := adb.ListApps(&database.ListAppOptions{
			ReadAppOptions: database.ReadAppOptions{
				Icon: false,
			},
			Plugin: app.Plugin,
			Owner:  &owner,
		})
		if err != nil {
			return "", "", err
		}
		if len(a) >= 1 {
			return "", "", errors.New("unique plugin app already exists")
		}
	}

	appv := App(*app.Plugin, owner, papp)
	assets.CopyStructIfPtrSet(&appv.Details, &app.Details)
	if c.DB.Type() == database.AdminType {
		// Copy all the fields from the app object
		assets.CopyStructIfPtrSet(appv, app)
	}

	aid, akey, err := adb.CreateApp(appv)
	if err != nil {
		if a.Config.Verbose {
			c.Log.Warnf("Creating plugin app failed: %s\n%s", err.Error(), appv)
		}
		return aid, akey, err
	}
	for skey, sv := range papp.Objects {
		// We perform the next stuff as admin
		if sv.AutoCreate == nil || *sv.AutoCreate == true {
			_, err := c.RequestBuffer(c, "POST", "/api/objects", AppObject(aid, skey, sv), map[string]string{"X-Heedy-As": "heedy"})
			if err != nil {
				adb.DelApp(aid)
				return "", "", err
			}
		}

	}

	return aid, akey, err
}
