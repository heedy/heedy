package connectordb

import (
	"connectordb/datastream"
	"connectordb/operator"
	"connectordb/pathwrapper"
	"connectordb/users"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// MetaLog logs device/stream creation and such to a meta log device.
// It must be used OVER an authoperator, since the authoperator specifies the user for
// which to write the metalog
type MetaLog struct {
	operator.Operator
	pathwrapper.Wrapper
	metalogID int64
	name      string
}

// AddMetaLog logs metalog information
func AddMetaLog(userID int64, o operator.Operator) (MetaLog, error) {
	usr, err := o.AdminOperator().ReadUserByID(userID)
	if err != nil {
		return MetaLog{}, err
	}

	ml := MetaLog{Operator: o, name: usr.Name}
	ml.Wrapper = pathwrapper.Wrap(ml)
	err = ml.checkcreate(usr.Name + "/meta/log")
	if err != nil {
		return ml, err
	}
	s, err := o.AdminOperator().ReadStream(usr.Name + "/meta/log")
	if err != nil {
		return ml, err
	}
	ml.metalogID = s.StreamID
	return ml, err
}

func (m MetaLog) checkcreate(path string) error {
	_, err := m.AdminOperator().ReadStream(path)
	if err != nil {
		return m.AdminOperator().CreateStream(path, `{"type": "object", "properties": {"cmd": {"type": "string"},"arg": {"type": "string"}},"required": ["cmd","arg"]}`)

	}
	return nil
}

func (m MetaLog) writeLog(cmd string, arg string) {
	dp := datastream.NewDatapoint()
	dp.Data = map[string]string{"cmd": cmd, "arg": arg}
	dp.Sender = m.Name()
	dpa := datastream.DatapointArray{dp}

	// First insert into this object's metalog
	err := m.AdminOperator().InsertStreamByID(m.metalogID, "", dpa, true)
	if err != nil {
		log.WithFields(log.Fields{"cmd": cmd, "arg": arg, "o": m.Name()}).Error("Metalog insert failed: ", err)
	}

	//Next, make sure that the owner of the arg also gets inserted if not this
	i := strings.Index(arg, "/")
	if i != -1 {
		arg = arg[:i]
	}
	//Make sure same user name - and if the user was deleted, then its devices don't exist
	if !strings.HasPrefix(arg+"/", m.name) && cmd != "DeleteUser" {
		m.checkcreate(arg + "/meta/log")
		//different user!
		err := m.AdminOperator().InsertStream(arg+"/meta/log", dpa, true)
		if err != nil {

		}
	}
}

func (m MetaLog) logUserID(userID int64, cmd string) {
	u, err := m.AdminOperator().ReadUserByID(userID)
	if err == nil {
		m.writeLog(cmd, u.Name)
	}
}

func (m MetaLog) logDeviceID(deviceID int64, cmd string) {
	d, err := m.AdminOperator().ReadDeviceByID(deviceID)
	if err != nil {
		return
	}
	u, err := m.AdminOperator().ReadUserByID(d.UserID)
	if err == nil {
		m.writeLog(cmd, u.Name+"/"+d.Name)
	}
}

func (m MetaLog) logStreamID(streamID int64, cmd string) {
	s, err := m.AdminOperator().ReadStreamByID(streamID)
	if err != nil {
		log.Errorf("Metalog couldn't find stream %d", streamID)
		return
	}
	d, err := m.AdminOperator().ReadDeviceByID(s.DeviceID)
	if err != nil {
		log.Errorf("Metalog couldn't find device %d", s.DeviceID)
		return
	}
	u, err := m.AdminOperator().ReadUserByID(d.UserID)
	if err == nil {
		m.writeLog(cmd, u.Name+"/"+d.Name+"/"+s.Name)
	} else {
		log.Errorf("Metalog couldn't find user %d", d.UserID)
	}
}

func (m MetaLog) CreateUser(name, email, password, role string, public bool) error {
	err := m.Operator.CreateUser(name, email, password, role, public)
	if err == nil {
		m.writeLog("CreateUser", name)
	}
	return err
}

func (m MetaLog) UpdateUserByID(userID int64, updates map[string]interface{}) error {
	err := m.Operator.UpdateUserByID(userID, updates)
	if err == nil {
		m.logUserID(userID, "UpdateUser")
	}
	return err
}
func (m MetaLog) DeleteUserByID(userID int64) error {
	u, _ := m.AdminOperator().ReadUserByID(userID)
	err := m.Operator.DeleteUserByID(userID)
	if err == nil && u != nil && u.Name != m.name {
		m.writeLog("DeleteUser", u.Name)
	}
	return err
}
func (m MetaLog) CreateDeviceByUserID(userID int64, devicename string, public bool) error {
	err := m.Operator.CreateDeviceByUserID(userID, devicename, public)
	if err == nil {
		d, err := m.AdminOperator().ReadDeviceByUserID(userID, devicename)
		if err == nil {
			m.logDeviceID(d.DeviceID, "CreateDevice")
		}
	}
	return err
}
func (m MetaLog) UpdateDeviceByID(deviceID int64, updates map[string]interface{}) error {
	err := m.Operator.UpdateDeviceByID(deviceID, updates)
	if err == nil {
		m.logDeviceID(deviceID, "UpdateDevice")
	}
	return err
}
func (m MetaLog) DeleteDeviceByID(deviceID int64) error {
	var u *users.User
	d, err := m.AdminOperator().ReadDeviceByID(deviceID)
	if err == nil {
		u, _ = m.AdminOperator().ReadUserByID(d.UserID)
	}
	err = m.Operator.DeleteDeviceByID(deviceID)
	if err == nil && u != nil {
		m.writeLog("DeleteDevice", u.Name+"/"+d.Name)
	}
	return err
}
func (m MetaLog) CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error {
	err := m.Operator.CreateStreamByDeviceID(deviceID, streamname, jsonschema)
	if err == nil {
		s, err := m.AdminOperator().ReadStreamByDeviceID(deviceID, streamname)
		if err == nil {
			m.logStreamID(s.StreamID, "CreateStream")
		}
	}
	return err
}
func (m MetaLog) UpdateStreamByID(streamID int64, updates map[string]interface{}) error {
	err := m.Operator.UpdateStreamByID(streamID, updates)
	if err == nil {
		m.logStreamID(streamID, "UpdateStream")
	}
	return err
}
func (m MetaLog) DeleteStreamByID(streamID int64, substream string) error {
	var d *users.Device
	var u *users.User
	s, err := m.AdminOperator().ReadStreamByID(streamID)
	if err == nil {
		d, err = m.AdminOperator().ReadDeviceByID(s.DeviceID)
		if err == nil {
			u, _ = m.AdminOperator().ReadUserByID(d.UserID)
		}
	}

	err = m.Operator.DeleteStreamByID(streamID, substream)
	if err == nil && u != nil {

		m.writeLog("DeleteStream", u.Name+"/"+d.Name+"/"+s.Name)
	}
	return err
}
