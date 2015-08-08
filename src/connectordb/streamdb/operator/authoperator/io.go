package authoperator

import (
	"connectordb/streamdb/datastream"
	"connectordb/streamdb/users"
)

//InsertStreamByID inserts into a stream given the stream's ID
func (o *AuthOperator) InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error {
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

	if !dev.RelationToStream(strm, sdevice).Gte(users.DEVICE) {
		return ErrPermissions
	}

	if strm.DeviceId != dev.DeviceId {
		//The writer is not the owner - we set the datastream.Datapoints' sender field
		for i := range data {
			data[i].Sender = o.Name()
		}

		//Since the writer is not the owner, if the stream is a downlink, write to the downlink substream
		if strm.Downlink {
			substream = "downlink"
		} else {
			//It isn't a downlink stream - check if current device is an admin
			if !dev.IsAdmin {
				//The device can't write the stream - this is to stop potential errors when users unknowingly
				//write to one of their devices' streams without understanding what that means
				return ErrPermissions
			}
		}

	} else {
		//The writer is reader. Ensure the sender field is empty
		for i := range data {
			data[i].Sender = ""
		}
	}
	return o.BaseOperator.InsertStreamByID(streamID, substream, data, restamp)
}
