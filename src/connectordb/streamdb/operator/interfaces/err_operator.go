package interfaces

import (
	"connectordb/streamdb/operator/messenger"
	"connectordb/streamdb/users"

	"connectordb/streamdb/datastream"

	"errors"

	"github.com/nats-io/nats"
)

var (
	BadOperatorErr  = errors.New("Invalid Operator")
	ErrOperatorName = " ERROR OPERATOR "
)

//BaseOperatorInterface are the functions which must be implemented in order to use Operator.
//If these functions are implemented, then the operator is complete, and all functionality
//of the database is available
type ErrOperator struct {
}

func (eo ErrOperator) Name() string {
	return ErrOperatorName
}

func (eo ErrOperator) User() (*users.User, error) {
	return nil, BadOperatorErr
}

func (eo ErrOperator) Device() (*users.Device, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadAllUsers() ([]users.User, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) CreateUser(username, email, password string) error {
	return BadOperatorErr
}
func (eo ErrOperator) ReadUser(username string) (*users.User, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadUserByID(userID int64) (*users.User, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) UpdateUser(modifieduser *users.User) error {
	return BadOperatorErr
}
func (eo ErrOperator) DeleteUserByID(userID int64) error {
	return BadOperatorErr
}

func (o ErrOperator) Login(username, password string) (*users.User, *users.Device, error) {
	return nil, nil, BadOperatorErr
}

func (eo ErrOperator) ReadAllDevicesByUserID(userID int64) ([]users.Device, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) CreateDeviceByUserID(userID int64, devicename string) error {
	return BadOperatorErr
}
func (eo ErrOperator) ReadDevice(devicepath string) (*users.Device, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadDeviceByID(deviceID int64) (*users.Device, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadDeviceByUserID(userID int64, devicename string) (*users.Device, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadDeviceByAPIKey(apikey string) (*users.Device, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) UpdateDevice(modifieddevice *users.Device) error {
	return BadOperatorErr
}
func (eo ErrOperator) DeleteDeviceByID(deviceID int64) error {
	return BadOperatorErr
}
func (eo ErrOperator) ReadAllStreamsByDeviceID(deviceID int64) ([]users.Stream, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) CreateStreamByDeviceID(deviceID int64, streamname, jsonschema string) error {
	return BadOperatorErr
}
func (eo ErrOperator) ReadStream(streampath string) (*users.Stream, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadStreamByID(streamID int64) (*users.Stream, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadStreamByDeviceID(deviceID int64, streamname string) (*users.Stream, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) UpdateStream(modifiedstream *users.Stream) error {
	return BadOperatorErr
}
func (eo ErrOperator) DeleteStreamByID(streamID int64, substream string) error {
	return BadOperatorErr
}
func (eo ErrOperator) LengthStreamByID(streamID int64, substream string) (int64, error) {
	return 0, BadOperatorErr
}
func (eo ErrOperator) TimeToIndexStreamByID(streamID int64, substream string, time float64) (int64, error) {
	return 0, BadOperatorErr
}
func (eo ErrOperator) InsertStreamByID(streamID int64, substream string, data datastream.DatapointArray, restamp bool) error {
	return BadOperatorErr
}
func (eo ErrOperator) GetStreamTimeRangeByID(streamID int64, substream string, t1 float64, t2 float64, limit int64) (datastream.DataRange, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) GetStreamIndexRangeByID(streamID int64, substream string, i1 int64, i2 int64) (datastream.DataRange, error) {
	return nil, BadOperatorErr
}

func (eo ErrOperator) SubscribeUserByID(userID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) SubscribeDeviceByID(deviceID int64, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) SubscribeStreamByID(streamID int64, substream string, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, BadOperatorErr
}

func (eo ErrOperator) CountUsers() (uint64, error) {
	return 0, BadOperatorErr
}

func (eo ErrOperator) CountStreams() (uint64, error) {
	return 0, BadOperatorErr
}
func (eo ErrOperator) CountDevices() (uint64, error) {
	return 0, BadOperatorErr
}
func (eo ErrOperator) ChangeDeviceAPIKey(devicepath string) (apikey string, err error) {
	return "", BadOperatorErr
}
func (eo ErrOperator) ChangeUserPassword(username, newpass string) error {
	return BadOperatorErr
}
func (eo ErrOperator) CreateDevice(devicepath string) error {
	return BadOperatorErr
}

func (eo ErrOperator) CreateStream(streampath, jsonschema string) error {
	return BadOperatorErr
}
func (eo ErrOperator) DeleteDevice(devicepath string) error {
	return BadOperatorErr
}
func (eo ErrOperator) DeleteStream(streampath string) error {
	return BadOperatorErr
}
func (eo ErrOperator) DeleteUser(username string) error {
	return BadOperatorErr
}
func (eo ErrOperator) ReadAllDevices(username string) ([]users.Device, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) ReadAllStreams(devicepath string) ([]users.Stream, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) SetAdmin(path string, isadmin bool) error {
	return BadOperatorErr
}
func (eo ErrOperator) GetStreamIndexRange(streampath string, i1 int64, i2 int64) (datastream.DataRange, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) InsertStream(streampath string, data datastream.DatapointArray, restamp bool) error {
	return BadOperatorErr
}
func (eo ErrOperator) GetStreamTimeRange(streampath string, t1 float64, t2 float64, limit int64) (datastream.DataRange, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) LengthStream(streampath string) (int64, error) {
	return 0, BadOperatorErr
}
func (eo ErrOperator) Subscribe(path string, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) SubscribeDevice(devpath string, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) SubscribeUser(username string, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) SubscribeStream(streampath string, chn chan messenger.Message) (*nats.Subscription, error) {
	return nil, BadOperatorErr
}
func (eo ErrOperator) TimeToIndexStream(streampath string, time float64) (int64, error) {
	return 0, BadOperatorErr
}
