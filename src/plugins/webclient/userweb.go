package webclient

import (
	"log"
	"net/http"
)

func getUserPage(se *SessionEnvironment) {
	log.Printf("getting user page")

	pageData := make(map[string]interface{})

	log.Printf("userdb: %p, user: %v\n", userdb, se.User)

	devices, err := userdb.ReadAllDevicesByUserID(se.User.UserId)
	pageData["devices"] = devices
	pageData["user"] = se.User
	pageData["flashes"] = se.Session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}
	se.Save()

	err = loginHomeTemplate.ExecuteTemplate(se.Writer, "root.html", pageData)
	if err != nil {
		http.Error(se.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func editUserPage(se *SessionEnvironment) {
	pageData := make(map[string]interface{})

	log.Printf("Editing %v", se.User.Name)
	devices, err := userdb.ReadAllDevicesByUserID(se.User.UserId)
	pageData["devices"] = devices
	pageData["user"] = se.User
	pageData["flashes"] = se.Session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}

	se.Save()
	err = userEditTemplate.ExecuteTemplate(se.Writer, "user_edit.html", pageData)
	if err != nil {
		http.Error(se.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func modifyUserAction(se *SessionEnvironment) {
	email := se.Request.PostFormValue("email")

	log.Printf("Modifying user %v, new email: %v", se.User.Name, email)

	// TODO someday change this to send a link to the user's email address
	// and when they click on it change the email (send them their email in the
	// url string encrypted so we don't need another table)
	if email != "" {
		se.User.Email = email

		log.Printf("email passed first check")
		err := se.Operator.UpdateUser(se.User)

		if err != nil {
			se.Session.AddFlash(err.Error())
		} else {
			se.Session.AddFlash("Settings Updated")
		}
	}
	se.Save()
	http.Redirect(se.Writer, se.Request, "/secure/edit", http.StatusTemporaryRedirect)
}

func modifyPasswordAction(se *SessionEnvironment) {

	p0 := se.Request.PostFormValue("current_password")
	p1 := se.Request.PostFormValue("password1")
	p2 := se.Request.PostFormValue("password2")

	log.Printf("Modifying user %v, new password: %v", se.User.Name, p1)

	if p1 == p2 && p1 != "" && se.User.ValidatePassword(p0) {
		se.User.SetNewPassword(p1)
		err := se.Operator.UpdateUser(se.User)

		if err != nil {
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

func deleteUserAction(se *SessionEnvironment) {
	p0 := se.Request.PostFormValue("password")
	log.Printf("Deleting user %v", se.User.Name)

	if se.User.ValidatePassword(p0) {
		err := se.Operator.DeleteUserByID(se.User.UserId)

		if err != nil {
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
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
