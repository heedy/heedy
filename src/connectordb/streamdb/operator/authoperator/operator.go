package authoperator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/operator/interfaces"
	"connectordb/streamdb/users"
	"errors"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrPermissions is thrown when an operator tries to do stuff it is not allowed to do
	ErrPermissions = errors.New("Access Denied")
	ErrBadPath     = errors.New("not a valid path")

	//UserlogSchema is the schema that is used for the userlog
	UserlogSchema = `{
						"type": "object",
						"properties": {
							"cmd": {"type": "string"},
							"arg": {"type": "string"}
						},
						"required": ["cmd","arg"]
					}`
)

//AuthOperator is the database proxy for a particular device.
//TODO: Operator does not auto-expire after time period
type AuthOperator struct {
	interfaces.BaseOperator //The operator which is used to interact with the database

	operatorPath string //The operator path is the string name of the operator
	devID        int64  //the id of the device - operatorPath is not enough, since name changes can happen in other threads

	userlogID int64 //The ID of the stream which provides the userlog
}

//Name is the path to the device underlying the operator
func (o *AuthOperator) Name() string {
	return o.operatorPath
}

//User returns the current user
func (o *AuthOperator) User() (usr *users.User, err error) {
	dev, err := o.BaseOperator.ReadDeviceByID(o.devID)
	if err != nil {
		return nil, err
	}
	return o.BaseOperator.ReadUserByID(dev.UserId)
}

//Device returns the current device
func (o *AuthOperator) Device() (*users.Device, error) {
	return o.BaseOperator.ReadDeviceByID(o.devID)
}

//Permissions returns whether the operator has permissions given by the string
func (o *AuthOperator) Permissions(perm users.PermissionLevel) bool {
	dev, err := o.Device()
	if err != nil {
		return false
	}
	return dev.GeneralPermissions().Gte(perm)
}

//UserLog logs the given command and argument to the special "log" stream for the user.
func (o *AuthOperator) UserLog(cmd string, arg string) error {
	data := make(map[string]string)
	data["cmd"] = cmd
	data["arg"] = arg

	dp := datastream.NewDatapoint()
	dp.Data = data
	dp.Sender = o.Name()
	err := o.BaseOperator.InsertStreamByID(o.userlogID, "", datastream.DatapointArray{dp}, true)
	if err != nil {
		log.WithFields(log.Fields{"cmd": cmd, "arg": arg, "o": o.Name()}).Error("Userlog insert failed: ", err)
	}

	return err
}

func (o *AuthOperator) getDevicePath(deviceID int64) (path string, err error) {
	dev, err := o.ReadDeviceByID(deviceID)
	if err != nil {
		return "", err
	}

	usr, err := o.ReadUserByID(dev.UserId)
	if err != nil {
		return "", err
	}
	return usr.Name + "/" + dev.Name, nil
}

func (o *AuthOperator) getStreamPath(streamID int64) (path string, err error) {
	s, err := o.ReadStreamByID(streamID)
	if err != nil {
		return "", err
	}
	devpath, err := o.getDevicePath(s.DeviceId)
	return devpath + "/" + s.Name, err
}

//UserLogDeviceID writes the userlog using a device ID
func (o *AuthOperator) UserLogDeviceID(deviceID int64, cmd string) error {
	devpath, err := o.getDevicePath(deviceID)
	if err != nil {
		return err
	}
	return o.UserLog(cmd, devpath)
}

//UserLogStreamID writes the userlog using a streamID
func (o *AuthOperator) UserLogStreamID(streamID int64, cmd string) error {
	spath, err := o.getStreamPath(streamID)
	if err != nil {
		return err
	}
	return o.UserLog(cmd, spath)
}

//CountAllUsers returns the total number of users contatined in the database
func (o *AuthOperator) CountUsers() (uint64, error) {
	if o.Permissions(users.ROOT) {
		return o.BaseOperator.CountUsers()
	}
	return 0, ErrPermissions
}

//CountAllDevices returns the total number of devices contatined in the database
func (o *AuthOperator) CountDevices() (uint64, error) {
	if o.Permissions(users.ROOT) {
		return o.BaseOperator.CountDevices()
	}
	return 0, ErrPermissions
}

//CountAllStreams returns the total number of streams contatined in the database
func (o *AuthOperator) CountStreams() (uint64, error) {
	if o.Permissions(users.ROOT) {
		return o.BaseOperator.CountStreams()
	}
	return 0, ErrPermissions
}
