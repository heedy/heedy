package webapp

// NOTE 2015-09-07 - Joseph Lewis
// this is left here for the future, we may eventually want to entirely disable
// modification of users and logins using user credentials through REST.
//
// This will become important as the number of third party applications grow and
// we will want to limit access of certain user things like payment info to
// only them.
//
// If/When we do that, we already have the infrastructure to enable all modifications
// through a direct iFrame/window back to us.
// This may also be useful when implementing an oauth2 provider.

// func init() {
// 	folderPath, _ := osext.ExecutableFolder()
// 	templatesPath := path.Join(folderPath, "templates")
// 	basePath := path.Join(templatesPath, "base.html")
//
// 	// Parses our templates relative to the template path including the base
// 	tMust := func(templateName string) *template.Template {
// 		tPath := path.Join(templatesPath, templateName)
// 		return template.Must(template.ParseFiles(tPath, basePath))
// 	}
//
// 	loginPageTemplate = tMust("login.html")
// }
// const(
// 	loginHomeTemplate *template.Template
// 	loginPageTemplate *template.Template
// )
//
// func getLogger(request *http.Request) *log.Entry {
// 	// if real-ip header exists, faddr=address (forwardedAddress) is logged
// 	// this fixes proxy routing
//
// 	fields := log.Fields{"addr": request.RemoteAddr, "uri": request.URL.String()}
// 	if realIP := request.Header.Get("X-Real-IP"); realIP != "" {
// 		fields["faddr"] = realIP
// 		if util.IsLocalhost(request.RemoteAddr) {
// 			delete(fields, "addr")
// 		}
// 	}
//
// 	return log.WithFields(fields)
// }
//
// type WebHandler func(se *SessionEnvironment, logger *log.Entry)
//
// // Display the login page
// func getLogin(writer http.ResponseWriter, request *http.Request) {
// 	logger := getLogger(request)
// 	logger.Debugf("Showing login page")
//
// 	se, err := NewSessionEnvironment(writer, request)
//
// 	// Don't log in somebody that's already logged in
// 	if err == nil && se.User != nil && se.Device != nil {
// 		http.Redirect(writer, request, "/secure/", http.StatusTemporaryRedirect)
// 		return
// 	}
//
// 	pageData := make(map[string]interface{})
//
// 	err = loginPageTemplate.ExecuteTemplate(writer, "login.html", pageData)
// 	if err != nil {
// 		http.Error(writer, err.Error(), http.StatusInternalServerError)
// 	}
// }
//
// // Process a login POST message
// func postLogin(writer http.ResponseWriter, request *http.Request) {
// 	logger := getLogger(request)
// 	userstr := request.PostFormValue("username")
// 	passstr := request.PostFormValue("password")
//
// 	usroperator, err := operator.NewUserLoginOperator(userdb, userstr, passstr)
// 	if err != nil {
// 		logger.WithFields(log.Fields{"op": "AUTH", "usr": userstr}).Warn(err.Error())
// 		http.Redirect(writer, request, "/login/?failed=true", http.StatusTemporaryRedirect)
// 		return
// 	}
// 	user, _ := usroperator.User()
// 	userdev, _ := usroperator.Device()
//
// 	logger = logger.WithField("usr", user.Name)
// 	logger.Info("Login")
//
// 	// Get a session. We're ignoring the error resulted from decoding an
// 	// existing session: Get() always returns a session, even if empty.
// 	session, _ := store.Get(request, sessionName)
// 	session.Values["authenticated"] = true
// 	session.Values["User"] = *user
// 	session.Values["Device"] = *userdev
// 	session.Values["OrigUser"] = *user
//
// 	if err := session.Save(request, writer); err != nil {
// 		logger.Error(err.Error())
// 		http.Error(writer, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	http.Redirect(writer, request, "/secure/", http.StatusTemporaryRedirect)
// }

//
// import (
// 	"net/http"
//
// 	log "github.com/Sirupsen/logrus"
// )
//
// func modifyUserAction(se *SessionEnvironment, logger *log.Entry) {
// 	logger = logger.WithField("op", "ModifyEmailAction")
// 	email := se.Request.PostFormValue("email")
//
// 	logger.Infof("new email: %v", email)
//
// 	// TODO someday change this to send a link to the user's email address
// 	// and when they click on it change the email (send them their email in the
// 	// url string encrypted so we don't need another table)
// 	if email != "" {
// 		se.User.Email = email
//
// 		logger.Debugf("email passed first check")
// 		err := se.Operator.UpdateUser(se.User)
//
// 		if err != nil {
// 			logger.Warn(err.Error())
// 			se.Session.AddFlash(err.Error())
// 		} else {
// 			se.Session.AddFlash("Settings Updated")
// 		}
// 	}
// 	se.Save()
// 	http.Redirect(se.Writer, se.Request, "/secure/edit", http.StatusTemporaryRedirect)
// }
//
// func modifyPasswordAction(se *SessionEnvironment, logger *log.Entry) {
// 	logger = logger.WithField("op", "ModifyPasswordAction")
//
// 	p0 := se.Request.PostFormValue("current_password")
// 	p1 := se.Request.PostFormValue("password1")
// 	p2 := se.Request.PostFormValue("password2")
//
// 	logger.Info() //Do not display password - just log that modification was requested
//
// 	if p1 == p2 && p1 != "" && se.User.ValidatePassword(p0) {
// 		se.User.SetNewPassword(p1)
// 		err := se.Operator.UpdateUser(se.User)
//
// 		if err != nil {
// 			logger.Warn(err.Error())
// 			se.Session.AddFlash(err.Error())
// 		} else {
// 			se.Session.AddFlash("Your password has been updated.")
// 		}
// 	} else {
// 		logger.Warn("Password Mismatch")
// 		se.Session.AddFlash("Your passwords did not match, try again.")
// 	}
//
// 	se.Save()
// 	http.Redirect(se.Writer, se.Request, "/secure/edit", http.StatusTemporaryRedirect)
// }
//
// func deleteUserAction(se *SessionEnvironment, logger *log.Entry) {
// 	logger = logger.WithField("op", "DeleteUserAction")
// 	p0 := se.Request.PostFormValue("password")
// 	logger.Info()
//
// 	if se.User.ValidatePassword(p0) {
// 		err := se.Operator.DeleteUserByID(se.User.UserId)
//
// 		if err != nil {
// 			logger.Warn(err.Error())
// 			se.Session.AddFlash(err.Error())
// 		} else {
// 			se.Session.AddFlash("Your user was deleted.")
// 		}
// 	} else {
// 		se.Session.AddFlash("Your passwords did not match, try again.")
// 	}
// 	se.Save()
// 	http.Redirect(se.Writer, se.Request, "/secure/edit", http.StatusTemporaryRedirect)
// }
