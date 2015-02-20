package web_client

import (
    "encoding/json"
    "encoding/xml"
    "flag"
    "github.com/gorilla/mux"
    "gopkg.in/mgo.v2/bson"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "strconv"
    "streamdb/dtypes"
    "time"
    "streamdb/users"
)

const (
    MaxUploadSizeBytes = 1024 * 100 // 100 Kb
)

var (
    errorGenericResult   = GenericResult{http.StatusInternalServerError, "An internal error occurred"}
    okGenericResult      = GenericResult{http.StatusOK, "Success"}
    emailExistsResult    = GenericResult{http.StatusConflict, "A user with this email already exists"}
    usernameExistsResult = GenericResult{http.StatusConflict, "A user with this username already exists"}
    badRequestResult     = GenericResult{http.StatusBadRequest, "Something in the URL is wrong, do the user, device, or stream doesn't exist?"}
    unauthorizedResult   = GenericResult{http.StatusUnauthorized, "The API key provided either doesn't exist our records or is disabled."}
    forbiddenResult      = GenericResult{http.StatusForbidden, "You do not have sufficient privliges to perform this action"}
    timedb               *dtypes.TypedDatabase

    ignoreBadApiKeys = flag.Bool("ignoreBadApiKeys", false, "Ignores bad api keys and processes all requests as superuser.")
    adminDevice = users.Device{Id:-1, Name:"userservice/internal", Enabled:true, Superdevice:true, OwnerId:-1, CanWrite:true, CanWriteAnywhere:true, UserProxy:true}
)

type UserServiceRequest struct {
    requestingDevice *users.Device // the device that originally requested this service, may not represent a "real" device in the database
    user *users.User // the user specified in the upload URL or nil if none
    device *users.Device // the device specified in the upload URL or nil if none
    stream *users.Stream // the stream in the upload or nil

    uploadText string // the body that was uploaded
    uploadType string // json, xml, bson, etc.
}

// The user service handler receives all variavbles and does some processing returning
// a plain object to be marshalled and returned.
type UserServiceHandler func(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{})

func decodeUrlParams(username, deviceid, streamid string) (*users.User, *users.Device, *users.Stream, error) {
    var user *users.User
    var device *users.Device
    var stream *users.Stream
    var reserr error
    var err error


    if username != "" {
        user, reserr = userdb.ReadUserByName(username)
    }

    if deviceid != "" {
        devicei, _ := strconv.Atoi(deviceid)
        device,  err = userdb.ReadDeviceById(int64(devicei))

        if reserr != nil {
            reserr = err
        }
    }

    if streamid != "" {
        streami, _ := strconv.Atoi(streamid)
        stream, err = userdb.ReadStreamById(int64(streami))

        if reserr != nil {
            reserr = err
        }
    }

    return user, device, stream, err
}

// Runs an authorization check on the api before calling the function
func apiAuth(h UserServiceHandler, requesterIsSuperdevice, userOwnsReqeuster, requesterIsDevice, requesterOwnsStream bool) http.HandlerFunc {

    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        var err error
        var requester *users.Device
        var user *users.User
        var device *users.Device
        var stream *users.Stream


        var resultcode int
        var result interface{}


        vars := mux.Vars(request)
        outputtype := vars["style"]

        // Do HTTP Authentication
        authUser, authPass, ok := request.BasicAuth()

        if ! ok && ! *ignoreBadApiKeys{
            writer.Header().Set("WWW-Authenticate", "Basic")
            resultcode, result = http.StatusUnauthorized, unauthorizedResult
            goto FinishOutput
        }

        user, device, stream, err = decodeUrlParams(vars["username"], vars["deviceid"], vars["streamid"])
        if err != nil {
            resultcode, result = http.StatusInternalServerError, errorGenericResult
            goto FinishOutput
        }


        // username means nothing to us
        if *ignoreBadApiKeys {
            requester = &adminDevice
        } else {
            requester, _ = userdb.ReadDeviceByApiKey(authPass)

            if authUser != "" {
                log.Printf("found username %s password %s", authUser, authPass)

                val, usr := userdb.ValidateUser(authUser, authPass)

                if val {
                    requester = usr.ToDevice()
                } else {
                    resultcode, result = http.StatusUnauthorized, unauthorizedResult
                    goto FinishOutput
                }

            } else {
                requester, err = userdb.ReadDeviceByApiKey(authPass)

                if err != nil {
                    resultcode, result = http.StatusUnauthorized, unauthorizedResult
                    goto FinishOutput
                }
            }
        }


        if ! requester.IsActive() {
            resultcode, result = http.StatusForbidden, forbiddenResult
            // TODO throttle/notify if too many weird requests.
            log.Printf("Denied inactive device request | device:%v", requester)
            goto FinishOutput
        }


        // Check for superdevices.
        if requesterIsSuperdevice && !requester.IsAdmin() {
            resultcode = http.StatusForbidden
            result = forbiddenResult

            // TODO throttle/notify if too many weird requests.
            log.Printf("Denied superdevice request | device:%v", requester)
            goto FinishOutput

        }

        if ! requester.IsAdmin() {
            if userOwnsReqeuster && (user == nil || user.Id != requester.OwnerId) ||
               requesterIsDevice && (device == nil || device.Id != requester.Id) ||
               requesterOwnsStream && (stream == nil || stream.OwnerId != requester.Id) {
                   resultcode, result = http.StatusForbidden, forbiddenResult
                   goto FinishOutput
            }
        }

        // TODO check for upload limits

        resultcode, result = h(request, requester, user, device, stream)

FinishOutput:
        var val []byte
        var ct string
        switch outputtype {
            case "json":
                val, err = json.Marshal(result)
                ct = "text/json"

                callbackname := request.FormValue("callback")
                if callbackname != "" {
                    ct = "application/javascript"
                    val = append([]byte(callbackname + "("), val...)
                    val = append(val, []byte(");")...)
                }

            case "xml":
                val, err = xml.Marshal(result)
                ct = "text/xml"
            case "bson":
                val, err = bson.Marshal(result)
                ct = "application/bson"
            default:
                val, err = json.Marshal(result)
                ct = "text/json"

        }

        if err != nil {
            log.Printf("Could not service read user request|err:%v", err)

            writer.WriteHeader(http.StatusInternalServerError)
            writer.Write([]byte(errorGenericResult.Message))
            return
        }

        writer.Header().Set("Content-Type", ct)
        writer.Header().Set("Content-Length", strconv.Itoa(len(val)))
        writer.WriteHeader(resultcode)

        writer.Write(val)

    })
}



func readAllUsers(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var result ReadUserResult
    result.Status = 200

    users, err := userdb.ReadAllUsers()

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    is_super := requestingDevice.IsAdmin()

    for _, u := range users {
        if is_super {
            result.Users = append(result.Users, *u)
        } else {
            result.Users = append(result.Users, u.ToClean())
        }
    }

    return http.StatusOK, result
}

func readUser(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var result ReadUserResult
    result.Status = 200

    if user == nil {
        return http.StatusBadRequest, badRequestResult
    }

    if requestingDevice.IsAdmin() || requestingDevice.IsOwnedBy(user) {
        result.Users = append(result.Users, *user)
    } else {
        result.Users = append(result.Users, user.ToClean())
    }

    return http.StatusOK, result
}


// Requires a user struct with the name, email, password filled in. The password
// should be plaintext.
func createUser(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {

    var userUpload users.User

    if err := readBodyUnmarshalAndError(request, &userUpload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    id, err := userdb.CreateUser(userUpload.Name, userUpload.Email, userUpload.Password)

    switch {
        case err == users.ERR_EMAIL_EXISTS:
            return http.StatusInternalServerError, emailExistsResult
        case err == users.ERR_USERNAME_EXISTS:
            return http.StatusInternalServerError, usernameExistsResult
        case err != nil:
            log.Printf("Could not service create user request |err:%v", err)
            return http.StatusInternalServerError, errorGenericResult
        default:
            return http.StatusOK, NewCreateSuccessResult(id)
    }
}

func readBodyUnmarshalAndError(request *http.Request, tofill interface{}) (error) {
    defer request.Body.Close()
    body, err := ioutil.ReadAll(io.LimitReader(request.Body, 1048576))

    if err != nil {
        return err
    }

    // BUG Joseph Lewis -- we should Unmarshal all kinds

    if err := json.Unmarshal(body, tofill); err != nil {
        return err
    }

    return nil
}


func updateUser(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var userUpload users.User

    if err := readBodyUnmarshalAndError( request, &userUpload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    if ! requestingDevice.IsAdmin() {
        userUpload.Admin = false
    }

    if err := userdb.UpdateUser(&userUpload); err != nil {
        log.Printf("Could not service update user request|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, okGenericResult
}


func deleteUser(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var userUpload users.User

    if err := readBodyUnmarshalAndError( request, &userUpload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    if err := userdb.DeleteUser(userUpload.Id); err != nil {
        log.Printf("Could not service update user request|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, okGenericResult
}


func readDevices(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    //var result ReadUserResult
    var result ReadDeviceResult
    result.Status = 200

    is_super := requestingDevice.IsAdmin()
    devs, err := userdb.ReadDevicesForUserId(user.Id)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    for _, d := range devs {

        if is_super {
            result.Devices = append(result.Devices, *d)
        } else {
            result.Devices = append(result.Devices, d.ToClean())
        }
    }

    return http.StatusOK, result
}

func readDevice(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    //var result ReadUserResult
    var result ReadDeviceResult
    result.Status = 200


    if requestingDevice.IsAdmin() || device.IsOwnedBy(user) {
        result.Devices = append(result.Devices, *device)
    } else {
        result.Devices = append(result.Devices, device.ToClean())
    }

    return http.StatusOK, result
}



func createDevice(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    //var result ReadUserResult
    var upload users.Device

    if err := readBodyUnmarshalAndError(request, &upload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    id, err := userdb.CreateDevice(upload.Name, user)

    switch {
        case err != nil:
            log.Printf("Could not service create user request |err:%v", err)
            return http.StatusInternalServerError, errorGenericResult
        default:
            return http.StatusOK, NewCreateSuccessResult(id)
    }
}


func updateDevice(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var upload users.Device
    upload = *device

    if err := readBodyUnmarshalAndError(request, &upload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    log.Printf("Modified device is now: %v", upload)

    if ! requestingDevice.IsAdmin() {
        upload.Superdevice = false
    }

    if err := userdb.UpdateDevice(&upload); err != nil {
        log.Printf("Could not service update device request|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, okGenericResult
}


func deleteDevice(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    err := userdb.DeleteDevice(device.Id)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    } else {
        return http.StatusOK, okGenericResult
    }
}


func readStream(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var result ReadStreamResult

    is_super := requestingDevice.IsAdmin()

    if is_super || requestingDevice.IsOwnedBy(user) {
        result.Streams = append(result.Streams, *stream)
    } else {
        result.Streams = append(result.Streams, stream.ToClean())
    }

    return http.StatusOK, result
}

func readAllStreams(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var result ReadStreamResult

    is_super := requestingDevice.IsAdmin()
    streams, err := userdb.ReadStreamsByDevice(device)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    for _, s := range streams {

        if is_super || requestingDevice.IsOwnedBy(user) {
            result.Streams = append(result.Streams, *s)
        } else {
            result.Streams = append(result.Streams, s.ToClean())
        }
    }

    return http.StatusOK, result
}


func createStream(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var result users.Stream

    if err := readBodyUnmarshalAndError(request, &result); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    // todo return id
    id, err := userdb.CreateStream(result.Name, result.Type, device)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, NewCreateSuccessResult(id)
}


func deleteStream(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {

    err := userdb.DeleteStream(stream.Id)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    } else {
        return http.StatusOK, okGenericResult
    }
}



func createDataKey(user *users.User, device *users.Device, stream *users.Stream) string {
    return user.Name + "/" + device.Name + "/" + stream.Name
}


func timeToUnixNano(timestamp string) (int64, error) {
    ts, err := strconv.ParseInt(timestamp,10,64)
    if err==nil {
        return ts,nil
    }
    var t time.Time
    err = t.UnmarshalText([]byte(timestamp))
    return t.UnixNano(), err
}


// Converts a time in ns to an iso standard string
func nanoToTimestamp(nano int64) string {

    str, err := time.Unix(0, int64(nano)).MarshalText()

    if err != nil {
        return "0000-00-00T00:00:00"
    } else {
        return string(str)
    }
}

func createDataPoint(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    if user == nil || device == nil || stream == nil {
        log.Printf("user, device, or stream is nil|usr:%v dev:%v stream:%v", user, device, stream)
        return http.StatusInternalServerError, errorGenericResult
    }

    dtype,ok := dtypes.GetType("text")
    if !ok {
        log.Printf("Unrecognized datatype")
        return http.StatusInternalServerError, errorGenericResult
    }
    result := dtype.New()

    if err := readBodyUnmarshalAndError(request, &result); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    if (!dtype.IsValid(result)) {
        log.Printf("Datapoint type invalid -> incorrect length/range")
        return http.StatusInternalServerError, errorGenericResult
    }

    err := timedb.InsertKey(createDataKey(user, device, stream), result, "")

    if err != nil {
        log.Printf("Timedb error while inserting|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, NewCreateSuccessResult(1)
}


// Reads the data between two indexes.
func readDataByTime(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var result DatapointResult

    vars := mux.Vars(request)
    si1 := vars["time1"]
    si2 := vars["time2"]


    ts1, err := timeToUnixNano(si1)
    if err != nil {
        log.Printf("Error converting timestamp|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }
    ts2, err := timeToUnixNano(si2)
    if err != nil {
        log.Printf("Error converting timestamp|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    log.Printf("Requesting ts (%v, %v] from %v", ts1, ts2, createDataKey(user, device, stream))

    dataReader := timedb.GetTimeRange(createDataKey(user, device, stream), "text", ts1, ts2)
    defer dataReader.Close()

    for {
        tmp:=dataReader.Next()

        if tmp==nil {
            break
        }

        result.Data = append(result.Data, tmp)
    }

    return http.StatusOK, result
}

// Reads the data between two indexes.
func readDataByIndex(request *http.Request, requestingDevice *users.Device, user *users.User, device *users.Device, stream *users.Stream) (int, interface{}) {
    var result DatapointResult

    vars := mux.Vars(request)
    si1 := vars["index1"]
    si2 := vars["index2"]


    i1, _ := strconv.Atoi(si1)
    i2, _ := strconv.Atoi(si2)


    log.Printf("Requesting data (%v, %v] from %v", i1, i2, createDataKey(user, device, stream))

    dataReader := timedb.GetIndexRange(createDataKey(user, device, stream),"text", uint64(i1), uint64(i2))
    defer dataReader.Close()

    for {
        tmp:=dataReader.Next()

        if tmp==nil {
            break
        }

        result.Data = append(result.Data, tmp)
    }

    return http.StatusOK, result
}


// Creates a subrouter available to
func GetSubrouter(udb *users.UserDatabase, tdb  *dtypes.TypedDatabase, subroutePrefix *mux.Router) {

    userdb = udb
    timedb = tdb

    usr, _ := userdb.ReadAllUsers()
    if len(usr) == 0 {
        log.Printf("No users found, auto ignoring bad api keys, all requests are superuser")
        *ignoreBadApiKeys = true
    }


    s := subroutePrefix.PathPrefix("/api/v1/{style}").Subrouter()

    s.HandleFunc("/user/", apiAuth(readAllUsers, false, false, false, false)).Methods("GET")
    s.HandleFunc("/user/", apiAuth(createUser,   true,  false, false, false)).Methods("POST")

    s.HandleFunc("/{username}/", apiAuth(readUser,   false, true, false, false)).Methods("GET")
    s.HandleFunc("/{username}/", apiAuth(updateUser, true, false, false, false)).Methods("PUT")
    s.HandleFunc("/{username}/", apiAuth(deleteUser, true, false, false, false)).Methods("DELETE")
    // username exists?
    //s.HandleFunc("/{AuthKey}/user/emailexists/{email}", apiAuth(deleteUser, true)).Methods("DELETE")
    //s.HandleFunc("/{AuthKey}/user/nameexists/{username}", apiAuth(deleteUser, true)).Methods("DELETE")

    // Requires params of username and password, username may be an email
    //s.HandleFunc("/{AuthKey}/user/authenticate/").Methods("POST")
    // validate user?

    s.HandleFunc("/{username}/device/", apiAuth(readDevices, false, true, false, false)).Methods("GET")
    s.HandleFunc("/{username}/device/", apiAuth(createDevice, false, true, false, false)).Methods("POST")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/", apiAuth(readDevice, false, true, false, false)).Methods("GET")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/", apiAuth(updateDevice, false, true, false, false)).Methods("PUT")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/", apiAuth(deleteDevice, true, false, false, false)).Methods("DELETE")

    s.HandleFunc("/{username}/{deviceid:[0-9]+}/stream/", apiAuth(readAllStreams, false, true, false, false)).Methods("GET")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/stream/", apiAuth(createStream, false, true, true, false)).Methods("POST")

    u := s.PathPrefix("/{username}/{deviceid:[0-9]+}/{streamid:[0-9]+}").Subrouter()

    u.HandleFunc("/", apiAuth(readStream,   false, false, false, false)).Methods("GET")
    u.HandleFunc("/", apiAuth(createStream, false, true, true, true)).Methods("POST")
    u.HandleFunc("/", apiAuth(deleteStream, true , false, false, false)).Methods("DELETE")

    u.HandleFunc("/point/i/{index1:[0-9]+}/{index2:[0-9]+}/", apiAuth(readDataByIndex, false, true, false, false)).Methods("GET")
    u.HandleFunc("/point/t/{time1}/{time2}/", apiAuth(readDataByTime, false, true, false, false)).Methods("GET")
    u.HandleFunc("/point/", apiAuth(createDataPoint, false, true, true, true)).Methods("POST")
}
