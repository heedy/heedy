package web_client

import (
	"html/template"
	"net/http"
	"path/filepath"
	"os"
	"os/exec"
    "streamdb/users"
	"streamdb"
    "github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"strconv"
)

var(
    userdb               *streamdb.UnifiedDB

	user_edit_template *template.Template
	login_home_template *template.Template
	device_info_template *template.Template
	firstrun_template *template.Template

	store = sessions.NewCookieStore([]byte("web-service-special-key"))
	firstrun bool
)


func getUserPage(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	pageData := make(map[string] interface{})

	devices, err := userdb.ReadDevicesForUserId(user.Id)
	pageData["devices"] = devices
	pageData["user"] = user
	pageData["flashes"] = session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}

	err = login_home_template.ExecuteTemplate(writer, "root.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func editUserPage(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	pageData := make(map[string] interface{})

	log.Printf("Editing %v", user.Name)
	devices, err := userdb.ReadDevicesForUserId(user.Id)
	pageData["devices"] = devices
	pageData["user"] = user
	pageData["flashes"] = session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}

	err = user_edit_template.ExecuteTemplate(writer, "user_edit.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func modifyUserAction(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	email := request.PostFormValue("email")

	log.Printf("Modifying user %v, new email: %v", user.Name, email)

	// TODO someday change this to send a link to the user's email address
	// and when they click on it change the email (send them their email in the
	// url string encrypted so we don't need another table)
	if email != "" {
		user.Email = email
		log.Printf("email passed first check")
		err := userdb.UpdateUserAs(user.ToDevice(), user)

		if err != nil {
			session.AddFlash(err.Error())
		} else {
			session.AddFlash("Settings Updated")
		}
	}

	http.Redirect(writer, request, "/secure/edit", http.StatusTemporaryRedirect)
}

func modifyPasswordAction(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	p0 := request.PostFormValue("current_password")
	p1 := request.PostFormValue("password1")
	p2 := request.PostFormValue("password2")

	log.Printf("Modifying user %v, new password: %v", user.Name, p1)


	if p1 == p2 && p1 != "" && user.ValidatePassword(p0) {
		user.SetNewPassword(p1)
		//err := userdb.UpdateUser(user)
		err := userdb.UpdateUserAs(user.ToDevice(), user)

		if err != nil {
			session.AddFlash(err.Error())
		} else {
			session.AddFlash("Your password has been updated.")
		}
	} else {
		session.AddFlash("Your passwords did not match, try again.")
	}

	http.Redirect(writer, request, "/secure/edit", http.StatusTemporaryRedirect)
}


func deleteUserAction(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	p0 := request.PostFormValue("password")

	log.Printf("Deleting user %v", user.Name)


	if user.ValidatePassword(p0) {
		user.SetNewPassword(p0)
		err := userdb.UpdateUserAs(user.ToDevice(), user)

		if err != nil {
			session.AddFlash(err.Error())
		} else {
			session.AddFlash("Your password has been updated.")
		}
	} else {
		session.AddFlash("Your passwords did not match, try again.")
	}

	http.Redirect(writer, request, "/secure/edit", http.StatusTemporaryRedirect)
}


func createDeviceAction(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	devname := request.PostFormValue("name")

	log.Printf("Creating device %v", devname)


	if devname != "" {
		devid, err := userdb.CreateDeviceAs(user.ToDevice(), devname, user)

		if err != nil {
			log.Printf(err.Error())
			session.AddFlash(err.Error())
		} else {
			session.AddFlash("Created Device")
			http.Redirect(writer, request, "/secure/device/" + strconv.Itoa(int(devid)), http.StatusTemporaryRedirect)
		}
	} else {
		session.AddFlash("You must enter a device name.")
	}

	http.Redirect(writer, request, "/secure/", http.StatusTemporaryRedirect)
}

func editDevicePage(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	vars := mux.Vars(request)
	devids := vars["id"]
	devid, _ := strconv.Atoi(devids)
	device, err := userdb.ReadDeviceById(int64(devid))

	if err != nil {
		session.AddFlash("Error getting device, maybe it was deleted?")
		goto redirect
	}

	device.Shortname 	 = request.PostFormValue("shortname")
	device.Enabled   	 = request.PostFormValue("enabled") == "checked"
	device.Superdevice   = request.PostFormValue("superdevice") == "checked"
	device.CanWrite		 = request.PostFormValue("canwrite") == "checked"
	device.CanWriteAnywhere	= request.PostFormValue("canwriteanywhere") == "checked"
	device.UserProxy     = request.PostFormValue("userproxy") == "checked"

	err = userdb.UpdateDeviceAs(user.ToDevice(), device)

	if err != nil {
		log.Printf(err.Error())
		session.AddFlash(err.Error())
	} else {
		session.AddFlash("Created Device")
	}

redirect:
	http.Redirect(writer, request, "/secure/device/" + devids, http.StatusTemporaryRedirect)
}



func firstRunHandler(writer http.ResponseWriter, r *http.Request) {
	pageData := make(map[string] interface{})

	pageData["alert"] = "All actions are admin, you should restart the server."


	err := firstrun_template.ExecuteTemplate(writer, "firstrun.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}


func getDevicePage(writer http.ResponseWriter, request *http.Request, user *users.User, session *sessions.Session) {
	pageData := make(map[string] interface{})

	vars := mux.Vars(request)
	devids := vars["id"]


	devid, _ := strconv.Atoi(devids)


	device, err := userdb.ReadDeviceById(int64(devid))
	pageData["device"] = device
	pageData["user"] = user
	pageData["flashes"] = session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting device."
	}

	streams, err := userdb.ReadStreamsByDevice(device)
	pageData["streams"] = streams

	if err != nil {
		pageData["alert"] = "Error getting device streams"
	}


	err = device_info_template.ExecuteTemplate(writer, "device_info.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
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



func Setup(subroutePrefix *mux.Router, udb *streamdb.UnifiedDB) {
    SetWdToExecutable()
	user_edit_template = template.Must(template.ParseFiles("./templates/user_edit.html", "./templates/base.html"))
	login_home_template = template.Must(template.ParseFiles("./templates/root.html", "./templates/base.html"))
	device_info_template = template.Must(template.ParseFiles("./templates/device_info.html", "./templates/base.html"))
	firstrun_template = template.Must(template.ParseFiles("./templates/firstrun.html", "./templates/base.html"))

	userdb = udb

	includepath, _ := filepath.Abs("./static/")
	log.Printf("Include path set to: %v", includepath)
	subroutePrefix.PathPrefix("/inc/").Handler(http.StripPrefix("/inc/", http.FileServer(http.Dir(includepath))))

	updateFirstrun()

	subroutePrefix.HandleFunc("/", authWrapper(getUserPage))
	subroutePrefix.HandleFunc("/secure/", authWrapper(getUserPage))
	subroutePrefix.HandleFunc("/secure/edit", authWrapper(editUserPage))
	subroutePrefix.HandleFunc("/secure/device/{id:[0-9]+}", authWrapper(getDevicePage))

	// CRUD user
	subroutePrefix.HandleFunc("/secure/user/action/modify", authWrapper(modifyUserAction))
	subroutePrefix.HandleFunc("/secure/user/action/changepass", authWrapper(modifyPasswordAction))
	subroutePrefix.HandleFunc("/secure/user/action/delete", authWrapper(deleteUserAction))

	// CRUD Device
	subroutePrefix.HandleFunc("/secure/device/action/create", authWrapper(createDeviceAction))
	subroutePrefix.HandleFunc("/secure/device/{id:[0-9]+}/action/edit", authWrapper(editDevicePage))

}
