// Package users provides an API for managing user information.
package users

import "testing"

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

	if k == j {
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


type ExpectedPemissions struct {
    in  string
    out PermissionLevel
	errExpected bool
}

var permissionsTest = []ExpectedPemissions{
	{"nobody", NOBODY, false},
	{"root", ROOT, false},
	{"user", USER, false},
	{"device", DEVICE, false},
	{"family", FAMILY, false},
	{"enabled", ENABLED, false},
	{"anybody", ANYBODY, false},
	{"", ANYBODY, true}}

func TestStrToPermissionLevel(t *testing.T) {
    for _, test := range permissionsTest {
		pl, err := strToPermissionLevel(test.in)

		if test.errExpected && err == nil || ! test.errExpected && err != nil {
			t.Errorf("Error failed for test %v", test.in)
		}

		if pl != test.out {
			t.Errorf("Wrong permission for %v, got %v expected %v", test.in, pl, test.out)
		}
    }
}

func TestDevicePermissions(t *testing.T) {
	var all Device
	all.IsAdmin = true
	all.Enabled = true
	all.CanWrite = true
	all.CanWriteAnywhere = true

	var none Device

	var onlyEnabled Device
	onlyEnabled.Enabled = true

	var disabledSuper Device
	disabledSuper.IsAdmin = true

	if none.IsActive() {
		t.Errorf("improper active check.")
	}

	if !onlyEnabled.IsActive() {
		t.Errorf("improper active check.")
	}

	if onlyEnabled.IsAdmin {
		t.Errorf("improper elevation of privliges.")
	}

	if !all.IsAdmin {
		t.Errorf("Correct admin was denied")
	}

	// WriteAllowed

	if none.WriteAllowed() {
		t.Errorf("Granted write to unprivliged")
	}

	if !all.WriteAllowed() {
		t.Errorf("Denied write to privliged device")
	}

	// WriteAnywhereAllowed

	if none.WriteAnywhereAllowed() {
		t.Errorf("Granted WriteAnywhereAllowed to unprivliged")
	}

	if !all.WriteAnywhereAllowed() {
		t.Errorf("Denied WriteAnywhereAllowed to privliged device")
	}

	// CanModifyUser

	if none.CanActAsUser {
		t.Errorf("Granted CanModifyUser to unprivliged")
	}
}

func TestGte(t *testing.T) {
	if ! NOBODY.Gte(ROOT) {
		t.Errorf("nobody should be gt root")
	}

	if ! ROOT.Gte(USER) {
		t.Errorf("root should be gt user")
	}

	if ! USER.Gte(DEVICE) {
		t.Errorf("user should be gt device")
	}

	if ! DEVICE.Gte(FAMILY) {
		t.Errorf("device should be gt family")
	}

	if ! FAMILY.Gte(ENABLED) {
		t.Errorf("family should be gt enabled")
	}

	if ! ENABLED.Gte(ANYBODY) {
		t.Errorf("enabled should be gt anybody")
	}

	if ! NOBODY.Gte(NOBODY) {
		t.Errorf("nobody should equal nobody")
	}
}

func TestGeneralPermissions(t *testing.T) {
	var d1 Device
	d1.IsAdmin = true
	d1.Enabled = true
	var d2 Device
	d2.Enabled = false
	var d3 Device
	d3.IsAdmin = true
	d3.Enabled = false
	var d4 Device
	d4.Enabled = true


	if d2.GeneralPermissions() != ANYBODY {
		t.Errorf("disabled devices should be anybody")
	}

	if d1.GeneralPermissions() != ROOT {
		t.Errorf("Enabled admin devices should be root")
	}

	if d3.GeneralPermissions() != ANYBODY {
		t.Errorf("disabled admins should be anybody")
	}

	if d4.GeneralPermissions() != ENABLED {
		t.Errorf("enabled non root should be enabled")
	}
}

func TestRelationToUser(t *testing.T) {
	u, dev, _, err := CreateUDS(testdb)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if dev.RelationToUser(nil) != ANYBODY {
		t.Errorf("nil user should be anybody")
	}

	dev.Enabled = false
	if dev.RelationToUser(u) != ANYBODY {
		t.Errorf("disabled devices should be anybody")
	}
	dev.Enabled = true

	dev.IsAdmin = true
	if dev.RelationToUser(u) != ROOT {
		t.Errorf("admin devices should have root")
	}
	dev.IsAdmin = false

	dev.CanActAsUser = true
	if dev.RelationToUser(u) != USER {
		t.Errorf("devices with same userid should be users")
	}
	dev.CanActAsUser = false

	if dev.RelationToUser(u) != DEVICE {
		t.Errorf("devices under a user should be a device")
	}

	dev.UserId = -1
	if dev.RelationToUser(u) != ANYBODY {
		t.Errorf("unrelated devices should be anybody")
	}
}



func TestRelationToDevice(t *testing.T) {
	u, dev, _, err := CreateUDS(testdb)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	d2, err := CreateTestDevice(testdb, u)
	if err != nil {
		t.Errorf(err.Error())
		return
	}


	if dev.RelationToDevice(nil) != ANYBODY {
		t.Errorf("nil dev should be anybody")
	}

	dev.Enabled = false
	if dev.RelationToDevice(d2) != ANYBODY {
		t.Errorf("disabled devices should be anybody")
	}
	dev.Enabled = true

	dev.IsAdmin = true
	if dev.RelationToDevice(d2) != ROOT {
		t.Errorf("admin devices should have root")
	}
	dev.IsAdmin = false

	dev.CanActAsUser = true
	if dev.RelationToDevice(d2) != USER {
		t.Errorf("devices with same userid should be users")
	}
	dev.CanActAsUser = false

	if dev.RelationToDevice(d2) != FAMILY {
		t.Errorf("devices under a user should be a family")
	}

	if dev.RelationToDevice(dev) != DEVICE {
		t.Errorf("Devices should be device with themselves")
	}

	dev.UserId = -1
	if dev.RelationToDevice(d2) != ENABLED {
		t.Errorf("unrelated devices should be enabled")
	}
}

func TestRelationToStream(t *testing.T) {
	_, dev, stream, err := CreateUDS(testdb)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if dev.RelationToStream(nil, dev) != ANYBODY {
		t.Errorf("nil stream")
	}

	if dev.RelationToStream(stream, nil) != ANYBODY {
		t.Errorf("nil parent")
	}

	dev.Enabled = false
	if dev.RelationToStream(stream, dev) != ANYBODY {
		t.Errorf("disabled dev")
	}
	dev.Enabled = true

	dev.IsAdmin = true
	if dev.RelationToStream(stream, dev) != ROOT {
		t.Errorf("root dev")
	}
	dev.IsAdmin = false

	dev.CanActAsUser = true
	if dev.RelationToStream(stream, dev) != USER {
		t.Errorf("root dev")
	}
	dev.CanActAsUser = false

	d2 := *dev
	d2.UserId = d2.UserId + 1
	d2.DeviceId = d2.DeviceId + 1
	if d2.RelationToStream(stream, dev) != ENABLED {
		t.Errorf("different user devices got %v", dev.RelationToStream(stream, &d2))
	}

	d2 = *dev
	d2.CanActAsUser = true
	d2.UserId += 1
	d2.DeviceId += 1
	if dev.RelationToStream(stream, &d2) == USER {
		t.Errorf("d2 can be user, but isn't parent")
	}

	if dev.RelationToStream(stream, dev) != DEVICE {
		t.Errorf("owner should be dev")
	}

}


/**



func (d *Device) RelationToDevice(device *Device) (PermissionLevel)  {
        // guards
        if device == nil || ! d.Enabled {
                return ANYBODY
        }

        // Permision Levels
        if d.IsAdmin {
                return ROOT
        }

        if d.UserId == device.UserId {
                if d.CanActAsUser {
                        return USER
                }

                if d.DeviceId == device.DeviceId {
                        return DEVICE
                }

                return FAMILY
        }


        if d.Enabled {
                return ENABLED
        }

        return ANYBODY
}


func (d *Device) RelationToStream(stream *Stream, streamParent *Device) (PermissionLevel)  {
        // guards
        if stream == nil || streamParent == nil || ! d.Enabled {
                return ANYBODY
        }

        // Permision Levels
        if d.IsAdmin {
                return ROOT
        }

        if d.CanActAsUser && d.UserId == streamParent.UserId {
                return USER
        }

        if d.DeviceId == stream.DeviceId {
                return DEVICE
        }

        if d.UserId == streamParent.UserId {
                return FAMILY
        }

        return ENABLED
}

**/
