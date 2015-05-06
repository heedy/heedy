package webclient

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"streamdb"
	"streamdb/users"
	"streamdb/util"
    "encoding/gob"

	"github.com/gorilla/mux"
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

	firstrun bool
	webOperator *streamdb.Operator
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
		log.Printf("Created session\n")

		if err != nil || se.User == nil || se.Device == nil {
			log.Printf("Error: %v, %v\n", err, se)
			http.Redirect(writer, request, "/login/", http.StatusTemporaryRedirect)
			return
		}
		log.Printf("Serving page\n")

		h(&se)
	})
}

// Display the login page
func getLogin(writer http.ResponseWriter, request *http.Request) {

	log.Printf("Showing login page\n")

	se, err := NewSessionEnvironment(writer, request)
	log.Printf("got session\n")

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
	log.Printf("logout\n")

	se, _ := NewSessionEnvironment(writer, request)
	se.Logoff()
	se.Save()

	http.Redirect(writer, request, "/login/", http.StatusTemporaryRedirect)
}

// Process a login POST message
func postLogin(writer http.ResponseWriter, request *http.Request) {
	userstr := request.PostFormValue("username")
	passstr := request.PostFormValue("password")

	log.Printf("Log in attempt: %v\n", userstr)

	user, userdev, err := userdb.Login(userstr, passstr)
	//_,_, err := userdb.Login(userstr, passstr)

	if err != nil {
		http.Redirect(writer, request, "/login/?failed=true", http.StatusTemporaryRedirect)
		return
	}

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
	util.SetWdToExecutable()

	userEditTemplate = template.Must(template.ParseFiles("./templates/user_edit.html", "./templates/base.html"))
	loginHomeTemplate = template.Must(template.ParseFiles("./templates/root.html", "./templates/base.html"))
	deviceInfoTemplate = template.Must(template.ParseFiles("./templates/device_info.html", "./templates/base.html"))
	firstrunTemplate = template.Must(template.ParseFiles("./templates/firstrun.html", "./templates/base.html"))
	addUserTemplate = template.Must(template.ParseFiles("./templates/newuser.html", "./templates/base.html"))
	loginPageTemplate = template.Must(template.ParseFiles("./templates/login.html", "./templates/base.html"))
}


func Setup(subroutePrefix *mux.Router, udb *streamdb.Database) {
	userdb = udb

	includepath, _ := filepath.Abs("./static/")
	log.Printf("Include path set to: %v", includepath)
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
	streamReadTemplate = template.Must(template.ParseFiles("./templates/stream.html", "./templates/base.html"))

	subroutePrefix.HandleFunc("/secure/stream/{id:[0-9]+}", authWrapper(readStreamPage))
	subroutePrefix.HandleFunc("/secure/stream/action/create/devid/{id:[0-9]+}", authWrapper(createStreamAction))
	subroutePrefix.HandleFunc("/secure/stream/{id:[0-9]+}/action/edit", authWrapper(editStreamAction))
	
}
