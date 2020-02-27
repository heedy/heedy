package plugins

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/events"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/sirupsen/logrus"
)

type Plugin struct {
	Name string
	Mux  *chi.Mux

	DB  *database.AdminDB
	Run *run.Manager

	// The root heedy server
	Server http.Handler

	EventRouter *events.Router
}

func NewPlugin(db *database.AdminDB, m *run.Manager, heedyServer http.Handler, pname string) (*Plugin, error) {
	p := &Plugin{
		DB:          db,
		Name:        pname,
		Run:         m,
		Server:      heedyServer,
		EventRouter: events.NewRouter(),
	}
	logrus.Debugf("Loading plugin '%s'", pname)

	return p, nil
}

func assetEventToEvent(ev assets.Event) events.Event {
	evt := events.Event{
		Event: ev.Event,
	}
	if ev.Type != nil {
		evt.Type = *ev.Type
	}
	if ev.Tags != nil {
		evt.Tags = &database.StringArray{}
		evt.Tags.Load(*ev.Tags)
	}
	evt.Plugin = ev.Plugin
	evt.Key = ev.Key

	return evt
}

func (p *Plugin) Start() error {

	pv := p.DB.Assets().Config.Plugins[p.Name]
	for rname, rv := range pv.Run {
		rv2 := rv // we need to pass a pointer to start, so need to create a new copy
		err := p.Run.Start(p.Name, rname, &rv2)
		if err != nil {
			logrus.Errorf("Failed to start %s:%s (%v)", p.Name, rname, err)
			p.Run.StopPlugin(p.Name)
			return err
		}
	}

	a := p.DB.Assets()

	psettings := a.Config.Plugins[p.Name]

	// Set up API forwards
	if psettings.Routes != nil && len(*psettings.Routes) > 0 {

		mux := chi.NewMux()

		for rname, redirect := range *psettings.Routes {
			logrus.Debugf("%s: Forwarding %s -> %s ", p.Name, rname, redirect)
			h, err := p.Run.GetHandler(p.Name, redirect)
			if err != nil {
				return err
			}
			err = run.Route(mux, rname, h)
			if err != nil {
				return err
			}

		}

		p.Mux = mux
	}

	// Set up events that are subscribed in the config with the "on" blocks

	for _, ev := range psettings.On {
		peh, err := NewPluginEventHandler(p, &ev)
		if err != nil {
			return err
		}
		evt := assetEventToEvent(ev)

		logrus.Debugf("%s: Forwarding event %s -> %s", p.Name, evt.String(), *ev.Post)
		p.EventRouter.Subscribe(evt, peh)
	}
	for cplugin, cv := range psettings.Apps {
		for _, ev := range cv.On {
			peh, err := NewPluginEventHandler(p, &ev)
			if err != nil {
				return err
			}
			cpn := p.Name + ":" + cplugin
			evt := assetEventToEvent(ev)
			evt.Plugin = &cpn
			logrus.Debugf("%s: Forwarding event %s -> %s", p.Name, evt.String(), *ev.Post)
			p.EventRouter.Subscribe(evt, peh)
		}
		for skey, sv := range cv.Objects {
			for _, ev := range sv.On {
				peh, err := NewPluginEventHandler(p, &ev)
				if err != nil {
					return err
				}
				cpn := p.Name + ":" + cplugin
				evt := assetEventToEvent(ev)
				evt.Plugin = &cpn
				if evt.Tags != nil {
					evt.Tags.Load(skey + " " + *ev.Tags)
				} else {
					evt.Tags = &database.StringArray{}
					evt.Tags.Load(skey)
				}
				logrus.Debugf("%s: Forwarding event %s -> %s", p.Name, evt.String(), *ev.Post)
				p.EventRouter.Subscribe(evt, peh)
			}
		}
	}
	// Attach the event router to the event system
	events.AddHandler(p.EventRouter)

	return nil
}

func (p *Plugin) AfterStart() error {

	a := p.DB.Assets()

	psettings := a.Config.Plugins[p.Name]

	// Make sure that all apps and objects that need to be auto-created are actually created

	for cname, cv := range psettings.Apps {
		pluginKey := p.Name + ":" + cname
		if cv.AutoCreate != nil && *cv.AutoCreate {
			// For each app
			// Check if the app exists for all users
			var res []string

			err := p.DB.DB.Select(&res, "SELECT username FROM users WHERE username NOT IN ('heedy', 'public', 'users') AND NOT EXISTS (SELECT 1 FROM apps WHERE owner=users.username AND apps.plugin=?);", pluginKey)
			if err != nil {
				return err
			}
			if len(res) > 0 {
				logrus.Debugf("%s: Creating '%s' app for all users", p.Name, pluginKey)

				// aaand how exactly do I achieve this?

				for _, uname := range res {

					_, _, err = p.DB.CreateApp(App(pluginKey, uname, cv))
					if err != nil {
						return err
					}
				}
			}
		}
		for skey, sv := range cv.Objects {
			if sv.AutoCreate == nil || *sv.AutoCreate == true {
				res := []string{}
				err := p.DB.DB.Select(&res, "SELECT id FROM apps WHERE plugin=? AND NOT EXISTS (SELECT 1 FROM objects WHERE app=apps.id AND key=?);", pluginKey, skey)
				if err != nil {
					return err
				}
				if len(res) > 0 {
					logrus.Debugf("%s: Creating '%s' object for all users with app '%s'", p.Name, skey, pluginKey)

					for _, cid := range res {
						s := AppObject(cid, skey, sv)
						_, err = run.Request(p.Server, "POST", "/api/objects", s, map[string]string{"X-Heedy-Key": p.Run.CoreKey})
						if err != nil {
							return err
						}
					}
				}
			}

		}
	}

	return nil
}

func (p *Plugin) OnUserCreate(username string) error {
	psettings := p.DB.Assets().Config.Plugins[p.Name]
	for cname, cv := range psettings.Apps {
		if cv.AutoCreate != nil && *cv.AutoCreate {
			// For each app

			pluginKey := p.Name + ":" + cname

			logrus.Debugf("%s: Creating '%s' app for user '%s'", p.Name, pluginKey, username)

			// aaand how exactly do I achieve this?

			cid, _, err := p.DB.CreateApp(App(pluginKey, username, cv))
			if err != nil {
				return err
			}

			for skey, sv := range cv.Objects {
				logrus.Debugf("%s: Creating '%s/%s' object for user '%s'", p.Name, pluginKey, skey, username)

				s := AppObject(cid, skey, sv)
				_, err = run.Request(p.Server, "POST", "/api/objects", s, map[string]string{"X-Heedy-Key": p.Run.CoreKey})
				if err != nil {
					return err
				}

			}
		}
	}
	return nil
}

func (p *Plugin) Close() error {
	events.RemoveHandler(p.EventRouter)
	return p.Run.StopPlugin(p.Name)
}
