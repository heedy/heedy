// Package users provides an API for managing user information.
package users

import "testing"

func TestUserToClean(t *testing.T) {
    uname := "MyName"
    var u User
    u.Name = uname
    u.Id = 50000

    v := u.ToClean()

    if u == v {
        t.Errorf("Not everything stripped from user got %v", v)
        return
    }

    var p User
    p.Name = uname

    if v != p {
        t.Errorf("Cleans do not match expecting %v, got %v", p, v)
        return
    }
}

func TestStreamToClean(t *testing.T) {
    s := Stream{Id: 11, Name:"Hello",
    Type:"blah",
    Ephemeral:true,
    Output:true,
    OwnerId:44}

    // no owner id
    r := Stream{Id: 11,
    Name:"Hello",
    Type:"blah",
    Ephemeral:true,
    Output:true}

    if s == r {
        t.Errorf("Init failed")
        return
    }

    v := s.ToClean()

    if v != r {
        t.Errorf("Clean does not equal test, got %v, expected %v", v, r)
        return
    }
}

func TestSetNewPassword(t *testing.T) {
    var j User
    var k User
    var l User

    j.SetNewPassword("monkey")
    k.SetNewPassword("password")
    l.PasswordSalt = "tmp"
    l.SetNewPassword("password")

    if j == k {
        t.Errorf("Setting password failed: %v vs %v", j, k)
        return
    }

    if k == l {
        t.Errorf("Salting Failed")
        return
    }

    j.SetNewPassword("password")

    if k != j {
        t.Errorf("Second Set Failed")
        return

    }
}


func TestAdmin(t *testing.T) {
    var j User
    j.Admin = true
    var k User
    k.Admin = false

    if k.IsAdmin() == true {
        t.Errorf("False positive admin")
        return
    }

    if j.IsAdmin() == false {
        t.Errorf("False negative admin")
        return
    }
}


func TestOwnsDevice(t *testing.T) {
    var j User
    j.Id = 1
    var k User
    k.Id = 42

    var d Device
    d.OwnerId = 42

    if j.OwnsDevice(&d) == true {
        t.Errorf("False positive on owns device")
        return
    }

    if k.OwnsDevice(&d) == false {
        t.Errorf("False negative on owns device")
        return
    }
}

func TestDeviceUserInteractions(t *testing.T) {
    var u User
    u.Admin = true
    u.Name = "cyberbob"
    u.Id = 42

    var u2 User

    pseudoU := u.ToDevice()

    var d Device

    if *pseudoU == d {
        t.Errorf("ToDevice returns defualt object.")
        return
    }

    if pseudoU.Name != u.Name {
        t.Errorf("Name not preserved.")
        return
    }

    if pseudoU.Shortname != u.Name {
        t.Errorf("Shortname not preserved.")
        return
    }

    if pseudoU.Superdevice != u.Admin {
        t.Errorf("Admin not preserved.")
        return
    }

    if pseudoU.OwnerId != u.Id {
        t.Errorf("Owner not preserved.")
        return
    }

    if *pseudoU.user != u {
        t.Errorf("User not set.")
        return
    }

    // The device side functions

    if pseudoU.IsUser() == false {
        t.Errorf("Determined not to be a user.")
        return
    }

    if *pseudoU.Unmask() != u {
        t.Errorf("Incorrect user pointer, expected %v got %v.", u, *pseudoU.Unmask() )
        return
    }

    if pseudoU.IsOwnedBy(&u2) {
        t.Errorf("Incorrect owner check.")
        return
    }

    if ! pseudoU.IsOwnedBy(&u) {
        t.Errorf("Incorrect owner check2.")
        return
    }
}



func TestDevicePermissions(t *testing.T) {
    var all Device
    all.Superdevice = true
    all.Enabled = true
    all.CanWrite = true
    all.CanWriteAnywhere = true
    all.UserProxy = true

    var none Device

    var onlyEnabled Device
    onlyEnabled.Enabled = true

    var disabledSuper Device
    disabledSuper.Superdevice = true

    if none.IsActive() {
        t.Errorf("improper active check.")
    }

    if ! onlyEnabled.IsActive() {
        t.Errorf("improper active check.")
    }

    if onlyEnabled.IsAdmin() {
        t.Errorf("improper elevation of privliges.")
    }

    if disabledSuper.IsAdmin() {
        t.Errorf("Granted admin to disabled device")
    }

    if ! all.IsAdmin() {
        t.Errorf("Correct admin was denied")
    }

    // WriteAllowed

    if none.WriteAllowed() {
        t.Errorf("Granted write to unprivliged")
    }

    if ! all.WriteAllowed() {
        t.Errorf("Denied write to privliged device")
    }

    // WriteAnywhereAllowed

    if none.WriteAnywhereAllowed() {
        t.Errorf("Granted WriteAnywhereAllowed to unprivliged")
    }

    if ! all.WriteAnywhereAllowed() {
        t.Errorf("Denied WriteAnywhereAllowed to privliged device")
    }


    // CanModifyUser

    if none.CanModifyUser() {
        t.Errorf("Granted CanModifyUser to unprivliged")
    }

    if ! all.CanModifyUser() {
        t.Errorf("Denied CanModifyUser to privliged device")
    }
}

func TestDeviceToClean(t *testing.T) {
    tmp := Device{1, "aa", "b", true, "base64", "a", true, 42, true, true, true, nil}
    cleaned := tmp.ToClean()

    if cleaned.Id != tmp.Id {
        t.Errorf("Id not preserved")
    }

    if cleaned.Name != tmp.Name {
        t.Errorf("Name not preserved")
    }

    if cleaned.Enabled != tmp.Enabled {
        t.Errorf("Enabled not preserved")
    }

    if cleaned.Icon_PngB64 != tmp.Icon_PngB64 {
        t.Errorf("Icon not preserved")
    }

    if cleaned.Shortname != tmp.Shortname {
        t.Errorf("Shortname not preserved")
    }

    if cleaned.Superdevice == tmp.Superdevice {
        t.Errorf("Admin leaked")
    }

    if cleaned.OwnerId == tmp.OwnerId {
        t.Errorf("Owner leaked")
    }

    if cleaned.CanWrite == tmp.CanWrite {
        t.Errorf("Write ability leaked")
    }

    if cleaned.CanWriteAnywhere == tmp.CanWriteAnywhere {
        t.Errorf("Write anywhere ability leaked")
    }

    if cleaned.UserProxy == tmp.UserProxy {
        t.Errorf("Proxy ability leaked")
    }
}
