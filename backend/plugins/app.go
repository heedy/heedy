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

func CreateApp(c *rest.Context, owner string, pluginKey string) (string, string, error) {
	if c.DB.Type() != database.UserType && c.DB.Type() != database.AdminType {
		return "", "", database.ErrAccessDenied("Only users can create apps")
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
	pk := strings.Split(pluginKey, ":")
	if len(pk) != 2 {
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}
	adb := c.DB.AdminDB()
	a := adb.Assets()

	p, ok := a.Config.Plugins[pk[0]]
	if !ok {
		return "", "", database.ErrBadQuery("invalid app plugin key")
	}

	app, ok := p.Apps[pk[1]]
	if !ok {
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

	if app.Unique != nil && *app.Unique {
		a, err := adb.ListApps(&database.ListAppOptions{
			ReadAppOptions: database.ReadAppOptions{
				Icon: false,
			},
			Plugin: &pluginKey,
			Owner:  &owner,
		})
		if err != nil {
			return "", "", err
		}
		if len(a) >= 1 {
			return "", "", errors.New("This unique plugin app already exists")
		}
	}

	aid, akey, err := adb.CreateApp(App(pluginKey, owner, app))
	if err != nil {
		if a.Config.Verbose {
			c.Log.Warnf("Creating plugin app failed: %s\n%s", err.Error(), App(pluginKey, owner, app))
		}
		return aid, akey, err
	}
	for skey, sv := range app.Objects {
		// We perform the next stuff as admin
		if sv.AutoCreate == nil || *sv.AutoCreate == true {
			_, err := c.Request(c, "POST", "/api/objects", AppObject(aid, skey, sv), map[string]string{"X-Heedy-As": "heedy"})
			if err != nil {
				adb.DelApp(aid)
				return "", "", err
			}
		}

	}

	return aid, akey, err
}
