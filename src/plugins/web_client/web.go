package web_client

import (
	"html/template"
	"net/http"
	"path/filepath"
	"os"
	"os/exec"
    "streamdb/users"
	"streamdb/timebatchdb"
    "github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
)

var(
    userdb               *users.UserDatabase
    timedb               *timebatchdb.Database
    templates *template.Template
	store = sessions.NewCookieStore([]byte("web-service-special-key"))
)


func generateMainPage(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	pageData := make(map[string] interface{})

	devices, err := userdb.ReadDevicesForUserId(user.Id)
	pageData["devices"] = devices
	pageData["user"] = user

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}

	//pageData["runs"] = ListDirectory("./autopilot/runs")
	//pageData["pilots"] = ListDirectory("./autopilot/pilots")
	//pageData["configurations"] = ListDirectory("./autopilot/configurations")

	//running := AutopilotRunning()
	//pageData["proc"] = running


	err = templates.ExecuteTemplate(writer, "root.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

}

func GenericCommand(command string, args ...string) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {

		//go func(){
			sp := exec.Command(command, args...)
			bytes, err := sp.CombinedOutput()
			//GenerateMainPage(w, "Running " + command)
		//}()

		err = templates.ExecuteTemplate(w, "command.html", string(bytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	}
}
func SetWdToExecutable() error {
	path, _ := exec.LookPath(os.Args[0])
    fp, _ := filepath.Abs(path)
    dir, _ := filepath.Split(fp)
    return os.Chdir(dir)
}

type WebHandler func(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session)

func authWrapper(h WebHandler) http.HandlerFunc {

    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        // Get a session. We're ignoring the error resulted from decoding an
        // existing session: Get() always returns a session, even if empty.
        session, _ := store.Get(request, "web-session")
        // Set some session values.

        // Do HTTP Authentication
        au, ap, _ := request.BasicAuth()

		var usr *users.User

        if au != "" && ap != "" {

			log.Printf("Got user %v pass %v", au, ap)

			var val bool
            val, usr = userdb.ValidateUser(au, ap)

            if ! val {
                writer.Header().Set("Content-Type", "text/plain")
                writer.Header().Set("WWW-Authenticate", "Basic")
                writer.WriteHeader(http.StatusUnauthorized)
                writer.Write([]byte("Username/Password wrong, please try again."))
                return
            }
        } else {
			writer.Header().Set("Content-Type", "text/plain")
			writer.Header().Set("WWW-Authenticate", "Basic")
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte("Username/Password wrong, please try again."))
			return
		}

        h(writer, request, usr, session)

    })
}



func Setup(subroutePrefix *mux.Router, udb *users.UserDatabase) {
    SetWdToExecutable()
	templates = template.Must(template.ParseGlob("./templates/*"))

	userdb = udb

	includepath, _ := filepath.Abs("./static/")
	log.Printf("Include path set to: %v", includepath)
	subroutePrefix.PathPrefix("/inc/").Handler(http.StripPrefix("/inc/", http.FileServer(http.Dir(includepath))))

    subroutePrefix.HandleFunc("/", authWrapper(generateMainPage))
}
