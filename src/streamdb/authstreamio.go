package streamdb

import "streamdb/users"

//LengthStream returns the length of the data in a stream
func (o *AuthOperator) LengthStream(streampath string) (int64, error) {
	_, err := o.ReadStream(streampath)
	if err != nil {
		return 0, err
	}
	//If we could read the stream, then we have permissions to get its length
	return o.Db.LengthStream(streampath)
}

//LengthStreamByID returns the length of the stream given its ID
func (o *AuthOperator) LengthStreamByID(streamID int64) (int64, error) {
	_, err := o.ReadStreamByID(streamID)
	if err != nil {
		return 0, err
	}
	//If we could read the stream, then we have permissions to get its length
	return o.Db.LengthStreamByID(streamID)
}

//InsertStream inserts into the stream
func (o *AuthOperator) InsertStream(streampath string, data []Datapoint) error {
	strm, err := o.ReadStream(streampath)
	if err != nil {
		return err
	}
	//No need to repeat the permissions stuff twice
	return o.InsertStreamByID(strm.StreamId, data)
}

//InsertStreamByID inserts into a stream given the stream's ID
func (o *AuthOperator) InsertStreamByID(streamID int64, data []Datapoint) error {
	strm, err := o.ReadStreamByID(streamID)
	if err != nil {
		return err
	}
	sdevice, err := o.ReadDeviceByID(strm.DeviceId)
	if err != nil {
		return err
	}

	dev, err := o.Device()
	if err != nil {
		return err
	}

	if !dev.RelationToStream(&strm.Stream, sdevice).Gte(users.DEVICE) {
		return ErrPermissions
	}

	if strm.DeviceId != dev.DeviceId {
		//The writer is not the owner - we set the datapoints' sender field
		for i := range data {
			data[i].Sender = o.Name()
		}
	} else {
		//The writer is reader. Ensure the sender field is empty
		for i := range data {
			data[i].Sender = ""
		}
	}
	return o.Db.InsertStreamByID(streamID, data)
}

//GetStreamTimeRange gets the time ragne
func (o *AuthOperator) GetStreamTimeRange(streampath string, t1 float64, t2 float64) (DatapointReader, error) {
	_, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.Db.GetStreamTimeRange(streampath, t1, t2)
}

//GetStreamTimeRangeByID gets the time range by ID
func (o *AuthOperator) GetStreamTimeRangeByID(streamID int64, t1 float64, t2 float64) (DatapointReader, error) {
	_, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	return o.Db.GetStreamTimeRangeByID(streamID, t1, t2)
}

//GetStreamIndexRange gets the index range by ID
func (o *AuthOperator) GetStreamIndexRange(streampath string, i1 int64, i2 int64) (DatapointReader, error) {
	_, err := o.ReadStream(streampath)
	if err != nil {
		return nil, err
	}
	return o.Db.GetStreamIndexRange(streampath, i1, i2)
}

//GetStreamIndexRangeByID gets an index range by ID
func (o *AuthOperator) GetStreamIndexRangeByID(streamID int64, i1 int64, i2 int64) (DatapointReader, error) {
	_, err := o.ReadStreamByID(streamID)
	if err != nil {
		return nil, err
	}
	return o.Db.GetStreamIndexRangeByID(streamID, i1, i2)
}
