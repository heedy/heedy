package webclient

import (
	"log"
	"net/http"
)


func getUserPage(srw *SessionResponseWriter) {
	writer := srw
	session := srw.Session()
	user, _, _ := srw.GetUserAndDevice()
	//user, userdevice, _ := srw.GetUserAndDevice()
	//operator, _ := userdb.GetOperatorForDevice(userdevice)
	pageData := make(map[string]interface{})

	devices, err := userdb.ReadDevicesForUserId(user.UserId)
	pageData["devices"] = devices
	pageData["user"] = user
	pageData["flashes"] = session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}

	err = loginHomeTemplate.ExecuteTemplate(writer, "root.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}


func editUserPage(srw *SessionResponseWriter) {
	writer := srw
	//request := srw.Request()
	session := srw.Session()
	//user, userdevice, _ := srw.GetUserAndDevice()
	user, _, _ := srw.GetUserAndDevice()
	//operator, _ := userdb.GetOperatorForDevice(userdevice)
	pageData := make(map[string]interface{})


	log.Printf("Editing %v", user.Name)
	devices, err := userdb.ReadDevicesForUserId(user.UserId)
	pageData["devices"] = devices
	pageData["user"] = user
	pageData["flashes"] = session.Flashes()

	if err != nil {
		pageData["alert"] = "Error getting devices."
	}

	err = userEditTemplate.ExecuteTemplate(writer, "user_edit.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func modifyUserAction(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	user, userdevice, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)


	email := request.PostFormValue("email")

	log.Printf("Modifying user %v, new email: %v", user.Name, email)

	// TODO someday change this to send a link to the user's email address
	// and when they click on it change the email (send them their email in the
	// url string encrypted so we don't need another table)
	if email != "" {
		originaluser := *user
		user.Email = email

		log.Printf("email passed first check")
		err := operator.UpdateUser(user, &originaluser)

		if err != nil {
			session.AddFlash(err.Error())
		} else {
			session.AddFlash("Settings Updated")
		}
	}

	http.Redirect(writer, request, "/secure/edit", http.StatusTemporaryRedirect)
}

func modifyPasswordAction(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	user, userdevice, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)

	p0 := request.PostFormValue("current_password")
	p1 := request.PostFormValue("password1")
	p2 := request.PostFormValue("password2")

	log.Printf("Modifying user %v, new password: %v", user.Name, p1)

	if p1 == p2 && p1 != "" && user.ValidatePassword(p0) {
		origuser := *user
		user.SetNewPassword(p1)
		err := operator.UpdateUser(user, &origuser)

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

func deleteUserAction(srw *SessionResponseWriter) {
	writer := srw
	request := srw.Request()
	session := srw.Session()
	user, userdevice, _ := srw.GetUserAndDevice()
	operator, _ := userdb.GetOperatorForDevice(userdevice)

	p0 := request.PostFormValue("password")

	log.Printf("Deleting user %v", user.Name)

	if user.ValidatePassword(p0) {
		err := operator.DeleteUser(user.UserId)

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

func newUserPage(writer http.ResponseWriter, r *http.Request) {
	pageData := make(map[string]interface{})

	err := addUserTemplate.ExecuteTemplate(writer, "newuser.html", pageData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
