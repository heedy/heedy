/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package authoperator

import (
	"connectordb/datastream"
	"connectordb/operator/interfaces"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//MetaLog logs the given command for the given path argument
func (o *AuthOperator) MetaLog(cmd string, arg string) error {
	dp := datastream.NewDatapoint()
	dp.Data = map[string]string{"cmd": cmd, "arg": arg}
	dp.Sender = o.Name()
	dpa := datastream.DatapointArray{dp}

	//First insert into this object's operator
	err := o.BaseOperator.InsertStreamByID(o.metalogID, "", dpa, true)
	if err != nil {
		log.WithFields(log.Fields{"cmd": cmd, "arg": arg, "o": o.Name()}).Error("Metalog insert failed: ", err)
	}

	//Next, make sure that the owner of the arg also gets inserted if not this
	i := strings.Index(arg, "/")
	if i != -1 {
		arg = arg[:i]
	}
	//Make sure same user name
	if !strings.HasPrefix(o.Name(), arg+"/") {
		//different user!
		interfaces.PathOperatorMixin{o.BaseOperator}.InsertStream(arg+"/meta/log", dpa, true)
	}

	return nil
}

//MetaLogDeviceID writes the userlog using a device ID
func (o *AuthOperator) MetaLogDeviceID(deviceID int64, cmd string) error {
	devpath, err := o.getDevicePath(deviceID)
	if err != nil {
		return err
	}
	return o.MetaLog(cmd, devpath)
}

//MetaLogStreamID writes the userlog using a streamID
func (o *AuthOperator) MetaLogStreamID(streamID int64, cmd string) error {
	spath, err := o.getStreamPath(streamID)
	if err != nil {
		return err
	}
	return o.MetaLog(cmd, spath)
}
