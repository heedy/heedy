package users

import(
	"errors"
	)

// The maximum size of a value in the key-value store
const (
	MaxKeyValueSizeBytes = 1024 * 4
)

var (
	MaximumSizeExceededError = errors.New("The inserted value is too big to store in the database")
	userKvTableInfo = keyValueTableInfo{"UserKeyValues", "UserId"}
	deviceKvTableInfo = keyValueTableInfo{"DeviceKeyValues", "DeviceId"}
	streamKvTableInfo = keyValueTableInfo{"StreamKeyValues", "StreamId"}
)


type keyValueTableInfo struct {
	TableName string
	IdField   string
}

// Generic create of a key value pair.
func (info* keyValueTableInfo) create(id int64, key, value string, userdb *UserDatabase) error {
	if len(value) > MaxKeyValueSizeBytes {
		return MaximumSizeExceededError
	}

	_, err := userdb.Exec("INSERT INTO " + info.TableName + " VALUES (?, ?, ?);", id, key, value)

	return err
}


// Generic update of a key value pair.
func (info* keyValueTableInfo) read(id int64, key string, userdb *UserDatabase, out interface{}) error {
	return userdb.Get(out, "SELECT * FROM " + info.TableName + " WHERE Key = ? AND " + info.IdField + " = ?;", key, id)
}

// Generic update of a key value pair.
func (info* keyValueTableInfo) update(id int64, key, value string, userdb *UserDatabase) error {
	if len(value) > MaxKeyValueSizeBytes {
		return MaximumSizeExceededError
	}

	_, err := userdb.Exec("UPDATE " + info.TableName + " SET Value = ? WHERE Key = ? AND " + info.IdField + " = ?;", value, key, id)

	return err
}

// Generic delete of a key value pair.
func (info* keyValueTableInfo) delete(id int64, key string, userdb *UserDatabase) error {
	_, err := userdb.Exec("DELETE FROM " + info.TableName + " WHERE Key = ? AND " + info.IdField + " = ?;", key, id)
	return err
}

// UserKeyValue stores key/value pairs for a particular user.
type UserKeyValue struct {
	UserId int64
	Key    string `modifiable:"root"`
	Value  string `modifiable:"user"`
}

// StreamKeyValue stores key/value pairs for a particular user as a given stream.
type StreamKeyValue struct {
	StreamId int64
	Key      string `modifiable:"root"`
	Value    string `modifiable:"device"`
}

// DeviceKeyValue stores device meta information in the KV store
type DeviceKeyValue struct {
	DeviceId int64
	Key      string `modifiable:"root"`
	Value    string `modifiable:"device"`
}




// CreateUserKeyValue creates a key value pair associated with a user
func (userdb *UserDatabase) CreateUserKeyValue(UserId int64, key, value string) error {
	return userKvTableInfo.create(UserId, key, value, userdb)
}

// CreateStreamKeyValue creates a key value pair associated with a stream
func (userdb *UserDatabase) CreateStreamKeyValue(StreamId int64, key, value string) error {
	return streamKvTableInfo.create(StreamId, key, value, userdb)
}

// CreateDeviceKeyValue creates a key value pair associated with a device
func (userdb *UserDatabase) CreateDeviceKeyValue(DeviceId int64, key, value string) error {
	return deviceKvTableInfo.create(DeviceId, key, value, userdb)
}


// ReadUserKeyValue reads a key value pair associated with a user
func (userdb *UserDatabase) ReadUserKeyValue(UserId int64, key string) (*UserKeyValue, error) {
	var kv UserKeyValue
	err := userKvTableInfo.read(UserId, key, userdb, &kv)
	return &kv, err
}

// ReadStreamKeyValue reads a key value pair associated with a stream
func (userdb *UserDatabase) ReadStreamKeyValue(StreamId int64, key string) (*StreamKeyValue, error) {
	var kv StreamKeyValue
	err := streamKvTableInfo.read(StreamId, key, userdb, &kv)
	return &kv, err
}

// ReadDeviceKeyValue reads a key value pair associated with a device
func (userdb *UserDatabase) ReadDeviceKeyValue(DeviceId int64, key string) (*DeviceKeyValue, error) {
	var kv DeviceKeyValue
	err := deviceKvTableInfo.read(DeviceId, key, userdb, &kv)
	return &kv, err
}


// UpdateUserKeyValue updates a key value pair associated with a user
func (userdb *UserDatabase) UpdateUserKeyValue(kv UserKeyValue) error {
	return userKvTableInfo.update(kv.UserId, kv.Key, kv.Value, userdb)
}

// UpdateStreamKeyValue updates a key value pair associated with a stream
func (userdb *UserDatabase) UpdateStreamKeyValue(kv StreamKeyValue) error {
	return streamKvTableInfo.update(kv.StreamId, kv.Key, kv.Value, userdb)
}

// UpdateDeviceKeyValue updates a key value pair associated with a device
func (userdb *UserDatabase) UpdateDeviceKeyValue(kv DeviceKeyValue) error {
	return deviceKvTableInfo.update(kv.DeviceId, kv.Key, kv.Value, userdb)
}

// DeleteUserKeyValue deletes a key value pair associated with a user
func (userdb *UserDatabase) DeleteUserKeyValue(kv UserKeyValue) error {
	return userKvTableInfo.delete(kv.UserId, kv.Key, userdb)
}

// DeleteStreamKeyValue deletes a key value pair associated with a stream
func (userdb *UserDatabase) DeleteStreamKeyValue(kv StreamKeyValue) error {
	return streamKvTableInfo.delete(kv.StreamId, kv.Key, userdb)
}

// DeleteDeviceKeyValue deletes a key value pair associated with a device
func (userdb *UserDatabase) DeleteDeviceKeyValue(kv DeviceKeyValue) error {
	return deviceKvTableInfo.delete(kv.DeviceId, kv.Key, userdb)
}




/**
// TODO see if we need this, it should probably be in dbutil

// StreamdbMeta holds information about the database itself, such as the version
// of streamdb
type StreamdbMeta struct {
	Key   string `modifiable:"root"`
	Value string `modifiable:"root"`
}


// CreateStreamdbMeta creates a streamdb meta tuple in the DB, errors if exists
// or on database error.
func (userdb *UserDatabase) CreateStreamdbMeta(key, value string) error {
	if len(value) > MaxKeyValueSizeBytes {
		return MaximumSizeExceededError
	}

	return userdb.Exec("INSERT INTO StreamdbMeta VALUES (?, ?);", key, value)
}

**/
