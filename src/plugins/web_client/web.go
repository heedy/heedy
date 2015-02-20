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
	"strconv"
)

var(
    userdb               *users.UserDatabase
    tdb               *timebatchdb.Database
    templates *template.Template
	store = sessions.NewCookieStore([]byte("web-service-special-key"))
	firstrun bool
)


func generateMainPage(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	pageData := make(map[string] interface{})

	devices, err := userdb.ReadDevicesForUserId(user.Id)
	pageData["devices"] = devices
	pageData["user"] = user

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}

	err = templates.ExecuteTemplate(writer, "root.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}




func firstRunHandler(writer http.ResponseWriter, r *http.Request) {
	pageData := make(map[string] interface{})

	pageData["alert"] = "All actions are admin, you should restart the server."


	err := templates.ExecuteTemplate(writer, "firstrun.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}


func generateDevicePage(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	pageData := make(map[string] interface{})

		vars := mux.Vars(request)
		devids := vars["id"]


		devid, _ := strconv.Atoi(devids)


		device, err := userdb.ReadDeviceById(int64(devid))
		pageData["device"] = device
		pageData["user"] = user

		if err != nil {
			pageData["alert"] = "Error getting device."
		}

		streams, err := userdb.ReadStreamsByDevice(device)
		pageData["streams"] = streams

		if err != nil {
			pageData["alert"] = "Error getting device streams"
		}


		err = templates.ExecuteTemplate(writer, "device_info.html", pageData)
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

		// all stuff redirects to firstrun
		if firstrun {
			if updateFirstrun() == true {
				firstRunHandler(writer, request)
				return
			}
		}

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


func updateFirstrun() bool {
	usr, _ := userdb.ReadAllUsers()
	firstrun = len(usr) == 0
	return firstrun
}



func Setup(subroutePrefix *mux.Router, udb *users.UserDatabase) {
    SetWdToExecutable()
	templates = template.Must(template.ParseGlob("./templates/*"))

	userdb = udb

	includepath, _ := filepath.Abs("./static/")
	log.Printf("Include path set to: %v", includepath)
	subroutePrefix.PathPrefix("/inc/").Handler(http.StripPrefix("/inc/", http.FileServer(http.Dir(includepath))))

	updateFirstrun()

	subroutePrefix.HandleFunc("/", authWrapper(generateMainPage))
	subroutePrefix.HandleFunc("/device/{id:[0-9]+}", authWrapper(generateDevicePage))
}
