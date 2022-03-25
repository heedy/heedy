package server

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/buildinfo"
	"github.com/heedy/heedy/backend/database"
	"github.com/lpar/gzipped/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var RunID string

func init() {
	RunID = strconv.FormatInt(buildinfo.StartTime.Unix(), 36)
}

type frontendPlugin struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type FrontendContext struct {
	User     *database.User                    `json:"user"`
	Settings map[string]map[string]interface{} `json:"settings"`
	Admin    bool                              `json:"admin"`
	Plugins  []frontendPlugin                  `json:"plugins"`
	Preload  []string                          `json:"preload"`
	Verbose  bool                              `json:"verbose"`
	DevMode  bool                              `json:"dev_mode"`
	RunID    string                            `json:"run_id"`
}

type aContext struct {
	User    *database.User `json:"user"`
	Request *CodeRequest   `json:"request"`
}

// withExists allows to use lpar/gzipped
type withExists struct {
	*afero.HttpFs
}

func (we withExists) Exists(name string) bool {
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return false
	}

	fullName := filepath.FromSlash(path.Clean("/" + name))
	_, err := we.Stat(fullName)
	return err == nil
}

func GetFrontendContext(ctx *rest.Context) (*FrontendContext, error) {
	var u *database.User
	var err error
	if _, ok := ctx.DB.(*database.UserDB); ok {
		u, err = ctx.DB.ReadUser(ctx.DB.ID(), &database.ReadUserOptions{
			Icon: true,
		})
		if err != nil {
			return nil, err
		}
	}

	cfg := assets.Config()

	frontendPlugins := make([]frontendPlugin, 0)
	preloads := make([]string, 0)
	if cfg.Frontend != nil {
		frontendPlugins = append(frontendPlugins, frontendPlugin{
			Name: "heedy",
			Path: *cfg.Frontend,
		})
		preloads = append(preloads, *cfg.Frontend)

	}
	if cfg.Preload != nil {
		preloads = append(preloads, (*cfg.Preload)...)
	}
	for _, p := range cfg.GetActivePlugins() {
		v, ok := cfg.Plugins[p]
		if !ok {
			return nil, errors.New("Failed to find plugin in configuration")
		}
		if v.Frontend != nil {
			frontendPlugins = append(frontendPlugins, frontendPlugin{
				Name: p,
				Path: *v.Frontend,
			})
			preloads = append(preloads, *v.Frontend)
		}
		if v.Preload != nil {
			preloads = append(preloads, (*v.Preload)...)
		}
	}

	if u == nil {
		return &FrontendContext{
			User:    nil,
			Admin:   false,
			Plugins: frontendPlugins,
			Preload: preloads,
			Verbose: cfg.Verbose,
			DevMode: buildinfo.DevMode,
			RunID:   RunID,
		}, nil
	}

	pref, err := ctx.DB.ReadUserSettings(*u.UserName)
	if err != nil {
		return nil, err
	}

	return &FrontendContext{
		User:     u,
		Settings: pref,
		Admin:    ctx.DB.AdminDB().Assets().Config.UserIsAdmin(*u.UserName),
		Plugins:  frontendPlugins,
		Preload:  preloads,
		Verbose:  cfg.Verbose,
		DevMode:  buildinfo.DevMode,
		RunID:    RunID,
	}, nil

}

func handleHTMLTemplate(fname string, fbytes []byte) (http.HandlerFunc, error) {
	fTemplate, err := template.New(fname).Parse(string(fbytes))
	if err != nil {
		return nil, err
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// Disallow clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "private, no-cache")

		ctx := rest.CTX(r)
		fCtx, err := GetFrontendContext(ctx)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		err = fTemplate.Execute(w, fCtx)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		return
	}, nil
}

func handleJSTemplate(fname string, fbytes []byte) (http.HandlerFunc, error) {
	// The templating engine understands javascript, but it must be within an html
	// <script> element. So we write the template like that, and then remove the script elements
	// when actually serving the page.
	const startTag = "<script>\n"
	const endTag = "\n</script>"
	scriptBytes := make([]byte, len(fbytes)+len(startTag)+len(endTag))
	copy(scriptBytes, startTag)
	copy(scriptBytes[len(startTag):], fbytes)
	copy(scriptBytes[len(scriptBytes)-len(endTag):], endTag)
	fTemplate, err := template.New(fname).Parse(string(scriptBytes))
	if err != nil {
		return nil, err
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-cache")
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")

		ctx := rest.CTX(r)
		fCtx, err := GetFrontendContext(ctx)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		var buf bytes.Buffer
		err = fTemplate.Execute(&buf, fCtx)
		if err != nil {
			rest.WriteJSONError(w, r, http.StatusInternalServerError, err)
		}
		// Remove the script tags and write the output
		outputBytes := buf.Bytes()
		outputBytes = outputBytes[len(startTag) : len(outputBytes)-len(endTag)]
		w.Header().Set("Content-Length", strconv.Itoa(len(outputBytes)))
		w.Write(outputBytes)
		return
	}, nil
}

func HandleTemplate(fname string, fbytes []byte) (http.HandlerFunc, error) {
	if strings.HasSuffix(fname, ".js") || strings.HasSuffix(fname, ".mjs") {
		return handleJSTemplate(fname, fbytes)
	}
	return handleHTMLTemplate(fname, fbytes)
}

// FrontendMux represents the frontend
func FrontendMux() (*chi.Mux, error) {
	mux := chi.NewMux()

	frontendFS := afero.NewBasePathFs(assets.Get().FS, "/public")

	logrus.Debug("frontend: preparing template index.html")
	fbytes, err := afero.ReadFile(frontendFS, "/index.html")
	if err != nil {
		return nil, err
	}
	h, err := HandleTemplate("index.html", fbytes)
	if err != nil {
		return nil, err
	}

	mux.Get("/", h)

	// And also prepare all the other templates from the directory

	files, err := afero.ReadDir(frontendFS, "")
	if err != nil {
		return nil, err
	}
	for i := range files {
		fname := files[i].Name()
		if !files[i].IsDir() && fname != "index.html" && fname != "auth.html" && !strings.HasPrefix(fname, "setup") {
			if strings.HasSuffix(fname, ".html") || strings.HasSuffix(fname, ".js") || strings.HasSuffix(fname, ".mjs") {
				logrus.Debug("frontend: preparing template ", fname)
				fbytes, err = afero.ReadFile(frontendFS, fname)
				if err != nil {
					return nil, err
				}
				h, err = HandleTemplate(fname, fbytes)
				if err != nil {
					return nil, err
				}
				if strings.HasSuffix(fname, ".html") {
					mux.Get("/"+fname[0:len(fname)-len(".html")], h)
				} else {
					mux.Get("/"+fname, h)
				}
			} else {
				mux.Get("/"+fname, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Cache-Control", "no-cache")
					fi, err := frontendFS.Stat(fname)
					if err != nil {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte("404 - Not Found"))
						return
					}
					fbytes, err := afero.ReadFile(frontendFS, fname)
					if err != nil {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte("404 - Not Found"))
						return
					}
					http.ServeContent(w, r, fname, fi.ModTime(), bytes.NewReader(fbytes))
				})
			}

		}
	}

	// Handles getting all assets other than the root webpage
	gzfs := gzipped.FileServer(withExists{afero.NewHttpFs(frontendFS)})
	if buildinfo.DevMode {
		mux.Mount("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// When in dev mode, we don't want any caching to happen. At all.
			w.Header().Set("Cache-Control", "no-cache")
			gzfs.ServeHTTP(w, r)
		}))
	} else {
		mux.Mount("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// We want caching, but we also want revalidation, so that heedy isn't left with old
			// code. We keep the response fresh for 10s, then keep it stale for a week.
			w.Header().Set("Cache-Control", "max-age=10,stale-while-revalidate=604800")
			gzfs.ServeHTTP(w, r)
		}))
	}

	return mux, nil
}
