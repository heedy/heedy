package webclient

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func getUserPage(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "GetUserPage")
	logger.Debugln()

	pageData := make(map[string]interface{})

	devices, err := webOperator.ReadAllDevicesByUserID(se.User.UserId)
	pageData["devices"] = devices
	pageData["user"] = se.User
	pageData["flashes"] = se.Session.Flashes()

	if err != nil {
		logger.Warn(err.Error())
		pageData["alert"] = "Error getting devices."
	}
	se.Save()

	err = loginHomeTemplate.ExecuteTemplate(se.Writer, "root.html", pageData)
	if err != nil {
		logger.Error(err.Error())
		http.Error(se.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func editUserPage(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "EditUserPage")
	logger.Debugln()
	pageData := make(map[string]interface{})

	devices, err := webOperator.ReadAllDevicesByUserID(se.User.UserId)
	pageData["devices"] = devices
	pageData["user"] = se.User
	pageData["flashes"] = se.Session.Flashes()

	if err != nil {
		logger.Warn(err.Error())
		pageData["alert"] = "Error getting devices."
	}

	se.Save()
	err = userEditTemplate.ExecuteTemplate(se.Writer, "user_edit.html", pageData)
	if err != nil {
		logger.Error(err.Error())
		http.Error(se.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func modifyUserAction(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "ModifyEmailAction")
	email := se.Request.PostFormValue("email")

	logger.Infof("new email: %v", email)

	// TODO someday change this to send a link to the user's email address
	// and when they click on it change the email (send them their email in the
	// url string encrypted so we don't need another table)
	if email != "" {
		se.User.Email = email

		logger.Debugf("email passed first check")
		err := se.Operator.UpdateUser(se.User)

		if err != nil {
			logger.Warn(err.Error())
			se.Session.AddFlash(err.Error())
		} else {
			se.Session.AddFlash("Settings Updated")
		}
	}
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/edit", http.StatusTemporaryRedirect)
}

func modifyPasswordAction(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "ModifyPasswordAction")

	p0 := se.Request.PostFormValue("current_password")
	p1 := se.Request.PostFormValue("password1")
	p2 := se.Request.PostFormValue("password2")

	logger.Info() //Do not display password - just log that modification was requested

	if p1 == p2 && p1 != "" && se.User.ValidatePassword(p0) {
		se.User.SetNewPassword(p1)
		err := se.Operator.UpdateUser(se.User)

		if err != nil {
			logger.Warn(err.Error())
			se.Session.AddFlash(err.Error())
		} else {
			se.Session.AddFlash("Your password has been updated.")
		}
	} else {
		logger.Warn("Password Mismatch")
		se.Session.AddFlash("Your passwords did not match, try again.")
	}

	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/edit", http.StatusTemporaryRedirect)
}

func deleteUserAction(se *SessionEnvironment, logger *log.Entry) {
	logger = logger.WithField("op", "DeleteUserAction")
	p0 := se.Request.PostFormValue("password")
	logger.Info()

	if se.User.ValidatePassword(p0) {
		err := se.Operator.DeleteUserByID(se.User.UserId)

		if err != nil {
			logger.Warn(err.Error())
			se.Session.AddFlash(err.Error())
		} else {
			se.Session.AddFlash("Your password has been updated.")
		}
	} else {
		se.Session.AddFlash("Your passwords did not match, try again.")
	}
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/edit", http.StatusTemporaryRedirect)
}

func newUserPage(writer http.ResponseWriter, r *http.Request) {
	pageData := make(map[string]interface{})

	err := addUserTemplate.ExecuteTemplate(writer, "newuser.html", pageData)
	if err != nil {
		getLogger(r).Errorln(err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
