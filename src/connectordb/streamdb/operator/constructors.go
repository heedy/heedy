package operator

/**
//NewAuthOperator creates a new authenticated operator,
func NewAuthOperator(db operator.BaseOperatorInterface, deviceID int64) (operator.PlainOperator, error) {
	dev, err := db.ReadDeviceByID(deviceID)
	if err != nil {
		return operator.Operator{}, err
	}
	usr, err := db.ReadUserByID(dev.UserId)
	if err != nil {
		return operator.Operator{}, err
	}

	userlogID, err := getUserLogStream(db, usr.UserId)
	if err != nil {
		return operator.Operator{}, err
	}

	return operator.{&AuthOperator{db, usr.Name + "/" + dev.Name, dev.DeviceId, userlogID}}, nil
}

//Returns the stream ID of the user log stream (and tries to create it if the stream does not exist)
func getUserLogStream(db operator.BaseOperatorInterface, userID int64) (streamID int64, err error) {
	o := operator.Operator{db}
	usr, err := o.ReadUserByID(userID)
	if err != nil {
		return 0, err
	}

	streamname := usr.Name + "/user/log"

	//Now attempt to go straight for the log stream
	logstream, err := o.ReadStream(streamname)
	if err != nil {
		//We had an error - try to create the stream (the user device is assumed to exist)
		err = o.CreateStream(streamname, UserlogSchema)
		if err != nil {
			return 0, err
		}

		//Now try to read the
		logstream, err = o.ReadStream(streamname)
	}
	return logstream.StreamId, err
}
**/
