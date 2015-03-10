package users

import("database/sql"
"errors"
"github.com/nu7hatch/gouuid"
)

// CreateDevice adds a device to the system given its owner and name.
// returns the last inserted id
func (userdb *UserDatabase) CreateDevice(Name string, OwnerId *User) (int64, error) {
    // guards
    if OwnerId == nil {
        return -1, ERR_INVALID_PTR
    }

    ApiKey, _ := uuid.NewV4()

    res, err := userdb.Db.Exec(CREATE_DEVICE_STMT,
            Name, ApiKey.String(), DEFAULT_ICON, OwnerId.Id)

    if err != nil {
        return -1, err
    }

    return res.LastInsertId()
}

// constructDeviceFromRow converts a SQL result to device by filling out a struct.
func constructDeviceFromRow(rows *sql.Rows, err error) (*Device, error) {

    result, err := constructDevicesFromRows(rows, err)

    if err != nil {
        return nil, err
    }

    if len(result) > 0 {
        return result[0], err
    }

    return nil, errors.New("No device supplied")
}

// constructDevicesFromRows constructs a series of devices
func constructDevicesFromRows(rows *sql.Rows, err error) ([]*Device, error) {
    out := []*Device{}

    if err != nil {
        return out, err
    }

    // defensive programming
    if rows == nil {
        return out, ERR_INVALID_PTR
    }

    defer rows.Close()
    for rows.Next() {
        u := new(Device)
        err := rows.Scan(
            &u.Id,
            &u.Name,
            &u.ApiKey,
            &u.Enabled,
            &u.Icon_PngB64,
            &u.Shortname,
            &u.Superdevice,
            &u.OwnerId,
            &u.CanWrite,
            &u.CanWriteAnywhere,
            &u.UserProxy)

        out = append(out, u)

        if(err != nil) {
            return out, err
        }
    }

    return out, nil
}

func (userdb *UserDatabase) ReadDevicesForUserId(Id int64) ([]*Device, error) {
    rows, err := userdb.Db.Query(SELECT_DEVICE_BY_USER_ID_STMT, Id)

    return constructDevicesFromRows(rows, err)
}

// ReadDeviceById selects the device with the given id from the database, returning nil if none can be found
func (userdb *UserDatabase) ReadDeviceById(Id int64) (*Device, error) {
    rows, err := userdb.Db.Query(SELECT_DEVICE_BY_ID_STMT, Id)
    return constructDeviceFromRow(rows, err)

}

// ReadDeviceByApiKey reads a device by an api key and returns it, it will be
// nil if an error was encountered and error will be set.
func (userdb *UserDatabase) ReadDeviceByApiKey(Key string) (*Device, error) {
    rows, err := userdb.Db.Query(SELECT_DEVICE_BY_API_KEY_STMT, Key)
    return constructDeviceFromRow(rows, err)
}

// UpdateDevice updates the given device in the database with all fields in the
// struct.
func (userdb *UserDatabase) UpdateDevice(device *Device) (error) {
if device == nil {
    return ERR_INVALID_PTR
}

_, err := userdb.Db.Exec(UPDATE_DEVICE_STMT,
    device.Name,
    device.ApiKey,
    device.Enabled,
    device.Icon_PngB64,
    device.Shortname,
    device.Superdevice,
    device.OwnerId,
    device.CanWrite,
    device.CanWriteAnywhere,
    device.UserProxy,
    device.Id)

    return err
}

// DeleteDevice removes a device from the system.
func (userdb *UserDatabase) DeleteDevice(Id int64) (error) {
    _, err := userdb.Db.Exec(DELETE_DEVICE_BY_ID_STMT, Id );
    return err
}
