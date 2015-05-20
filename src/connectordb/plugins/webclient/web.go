package webclient

import (
	"connectordb/streamdb"
	"connectordb/streamdb/operator"
	"connectordb/streamdb/users"
	"encoding/gob"
	"html/template"
	"net/http"
	"path"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/kardianos/osext"
)

var (
	userdb *streamdb.Database

	userEditTemplate   *template.Template
	loginHomeTemplate  *template.Template
	deviceInfoTemplate *template.Template
	firstrunTemplate   *template.Template
	streamReadTemplate *template.Template
	addUserTemplate    *template.Template
	loginPageTemplate  *template.Template

	firstrun    bool
	webOperator *operator.Operator
)

func init() {
	gob.Register(users.User{})
	gob.Register(users.Device{})
}

/**
func internalServerError(err error) {

}
**/

type WebHandler func(se *SessionEnvironment)

func authWrapper(h WebHandler) http.HandlerFunc {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		log.Printf("Doing login\n")
		se, err := NewSessionEnvironment(writer, request)
		log.Debugf("Created session\n")

		if err != nil || se.User == nil || se.Device == nil {
			log.Errorf("Error: %v, %v\n", err, se)
			http.Redirect(writer, request, "/login/", http.StatusTemporaryRedirect)
			return
		}
		log.Debugf("Serving page\n")

		h(&se)
	})
}

// Display the login page
func getLogin(writer http.ResponseWriter, request *http.Request) {

	log.Printf("Showing login page")

	se, err := NewSessionEnvironment(writer, request)
	log.Debugf("got session")

	// Don't log in somebody that's already logged in
	if err == nil && se.User != nil && se.Device != nil {
		http.Redirect(writer, request, "/secure/", http.StatusTemporaryRedirect)
		return
	}

	pageData := make(map[string]interface{})

	err = loginPageTemplate.ExecuteTemplate(writer, "login.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

// Display the login page
func getLogout(writer http.ResponseWriter, request *http.Request) {
	log.Printf("logout")

	se, _ := NewSessionEnvironment(writer, request)
	se.Logoff()
	se.Save()

	http.Redirect(writer, request, "/login/", http.StatusTemporaryRedirect)
}

// Process a login POST message
func postLogin(writer http.ResponseWriter, request *http.Request) {
	userstr := request.PostFormValue("username")
	passstr := request.PostFormValue("password")

	log.Printf("Log in attempt: %v", userstr)

	usroperator, err := userdb.LoginOperator(userstr, passstr)
	if err != nil {
		http.Redirect(writer, request, "/login/?failed=true", http.StatusTemporaryRedirect)
		return
	}
	user, _ := usroperator.User()
	userdev, _ := usroperator.Device()

	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := store.Get(request, sessionName)
	session.Values["authenticated"] = true
	session.Values["User"] = *user
	session.Values["Device"] = *userdev
	session.Values["OrigUser"] = *user

	if err := session.Save(request, writer); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(writer, request, "/secure/", http.StatusTemporaryRedirect)
}

func init() {
	folderPath, _ := osext.ExecutableFolder()
	templatesPath := path.Join(folderPath, "templates")
	basePath := path.Join(templatesPath, "base.html")

	// Parses our templates relative to the template path including the base
	// everything needs
	tMust := func(templateName string) *template.Template {
		tPath := path.Join(templatesPath, templateName)
		return template.Must(template.ParseFiles(tPath, basePath))
	}

	userEditTemplate = tMust("user_edit.html")
	loginHomeTemplate = tMust("root.html")
	deviceInfoTemplate = tMust("device_info.html")
	firstrunTemplate = tMust("firstrun.html")
	addUserTemplate = tMust("newuser.html")
	loginPageTemplate = tMust("login.html")
}

func Setup(subroutePrefix *mux.Router, udb *streamdb.Database) {
	userdb = udb
	folderPath, _ := osext.ExecutableFolder()
	includepath := path.Join(folderPath, "static")
	log.Debugf("Include path set to: %v", includepath)
	subroutePrefix.PathPrefix("/inc/").Handler(http.StripPrefix("/inc/", http.FileServer(http.Dir(includepath))))

	subroutePrefix.HandleFunc("/login/", http.HandlerFunc(getLogin))
	subroutePrefix.HandleFunc("/login/action/login", http.HandlerFunc(postLogin))
	subroutePrefix.HandleFunc("/login/action/logoff", http.HandlerFunc(getLogout))

	subroutePrefix.HandleFunc("/", http.HandlerFunc(getLogin))
	subroutePrefix.HandleFunc("/secure/", authWrapper(getUserPage))
	subroutePrefix.HandleFunc("/secure/edit", authWrapper(editUserPage))

	subroutePrefix.HandleFunc("/newuser/", newUserPage)

	// CRUD user
	subroutePrefix.HandleFunc("/secure/user/action/modify", authWrapper(modifyUserAction))
	subroutePrefix.HandleFunc("/secure/user/action/changepass", authWrapper(modifyPasswordAction))
	subroutePrefix.HandleFunc("/secure/user/action/delete", authWrapper(deleteUserAction))

	// CRUD Device
	subroutePrefix.HandleFunc("/secure/device/{id:[0-9]+}", authWrapper(getDevicePage))
	subroutePrefix.HandleFunc("/secure/device/action/create", authWrapper(createDeviceAction))
	subroutePrefix.HandleFunc("/secure/device/{id:[0-9]+}/action/edit", authWrapper(editDevicePage))

	// CRUD Stream
	streamReadTemplate = template.Must(template.ParseFiles(path.Join(folderPath, "./templates/stream.html"), path.Join(folderPath, "./templates/base.html")))

	subroutePrefix.HandleFunc("/secure/stream/{id:[0-9]+}", authWrapper(readStreamPage))
	subroutePrefix.HandleFunc("/secure/stream/action/create/devid/{id:[0-9]+}", authWrapper(createStreamAction))
	subroutePrefix.HandleFunc("/secure/stream/{id:[0-9]+}/action/edit", authWrapper(editStreamAction))

}
