package users

import (
    "github.com/gorilla/mux"
    "gopkg.in/mgo.v2/bson"
    "net/http"
    "log"
    "flag"
    "encoding/json"
    "io"
    "io/ioutil"
    "strconv"
    "encoding/xml"
    )

type key int

var (
    UNAUTHORIZED_MESSAGE = []byte("The API key provided either doesn't exist our records or is disabled.")
    ERROR_MESSAGE = []byte("An internal error occurred, we'll get right on that, sorry!")
    BAD_REQUEST_MESSAGE = []byte("Something in the URL is wrong, do the user, device, or stream doesn't exist?")
    FORBIDDEN_MESSAGE = []byte("You do not have sufficient privliges to perform this action")
    ignoreBadApiKeys = flag.Bool("ignoreBadApiKeys", false, "Ignores bad api keys and processes all requests as superuser.")
    errorGenericResult = GenericResult{http.StatusInternalServerError, "An internal error occurred"}
    okGenericResult = GenericResult{http.StatusOK, "Success"}
    emailExistsResult = GenericResult{http.StatusConflict, "A user with this email already exists"}
    usernameExistsResult = GenericResult{http.StatusConflict, "A user with this username already exists"}
    badRequestResult = GenericResult{http.StatusBadRequest, "Something in the URL is wrong, do the user, device, or stream doesn't exist?"}
    userdb *UserDatabase
)

const (
    REQUEST_DEVICE_IS_SUPERUSER key = 0
)

type GenericResult struct {
    Status int  // An HTTP status code
    Message string  // Extra data needed to pass along
}


type ReadUserResult struct {
    Users []CleanUser
    Unsanitized []User
    GenericResult
}

type ReadDeviceResult struct {
    Devices []CleanDevice
    Unsanitized []Device
    GenericResult
}

type ReadStreamResult struct {
    Streams []CleanStream
    Unsanitized []Stream
    GenericResult
}

type CreateSuccessResult struct {
    Id int64
    GenericResult
}

func NewCreateSuccessResult(id int64) CreateSuccessResult {
    var res CreateSuccessResult
    res.Id = id
    res.Message = "Success"
    res.Status = http.StatusOK

    return res
}

func (result GenericResult) writeToHttp(writer http.ResponseWriter) {

    if result.Status == 0 {
        result.Status = http.StatusOK
    }

    val, err := json.Marshal(result)

    if err != nil {
        log.Printf("Could not marshal data structure to json|err:%v result:%v", err, result)

        writer.WriteHeader(http.StatusInternalServerError)
        writer.Write(ERROR_MESSAGE)
        return
    }

    writer.WriteHeader(result.Status)
    writer.Write(val)
}

// Tests to see if this result was a "success" or not.
func (r GenericResult) IsSuccess() bool {
    return r.Status == 200 || r.Status == 204 // HTTP success code for ok
}

func getUserFromRequest(request *http.Request) (*User, error) {
    vars := mux.Vars(request)
    auth := vars["username"]
    return userdb.ReadUserByName(auth)
}

// The user service handler receives all variavbles and does some processing returning
// a plain object to be marshalled and returned.
type UserServiceHandler func(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{})

// Runs an authorization check on the api before calling the function
func apiAuth(h UserServiceHandler, requesterIsSuperdevice, userOwnsReqeuster, requesterIsDevice, requesterOwnsStream bool) http.HandlerFunc {

    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {


        // Do HTTP Authentication
        username, password, ok := request.BasicAuth()

        if ! ok && ! *ignoreBadApiKeys{
            writer.Header().Set("WWW-Authenticate", "Basic")
            writer.WriteHeader(http.StatusUnauthorized)
            writer.Write(UNAUTHORIZED_MESSAGE)

            return
        }

        // username means nothing to us
        var err error
        var requester *Device

        if *ignoreBadApiKeys {
            requester = new(Device)
            requester.Superdevice = true
            requester.Enabled = true
            requester.Shortname = "userservice/superadmin" // can't occur naturally
        } else {
            requester, err = userdb.ReadDeviceByApiKey(password)

            if username != "" {
                log.Printf("found username %s password %s", username, password)

                val, usr := userdb.ValidateUser(username, password)

                if val {
                    log.Print("correct user!")

                    requester = new(Device)
                    requester.Superdevice = usr.IsAdmin()
                    requester.Enabled = true
                    requester.Shortname = username
                    requester.Name = username
                    requester.OwnerId = usr.Id
                    requester.Id = -1
                } else {
                    writer.WriteHeader(http.StatusUnauthorized)
                    writer.Write(UNAUTHORIZED_MESSAGE)
                    return
                }

            } else {
                requester, err = userdb.ReadDeviceByApiKey(password)

                if err != nil {
                    writer.WriteHeader(http.StatusUnauthorized)
                    writer.Write(UNAUTHORIZED_MESSAGE)
                    return
                }
            }
        }


        // Check for superdevices.
        if requesterIsSuperdevice && !requester.isAdmin() {
            writer.WriteHeader(http.StatusForbidden)
            writer.Write(FORBIDDEN_MESSAGE)

            // TODO throttle/notify if too many weird requests.
            log.Printf("Denied superdevice request | device:%v", requester)
            return

        }

        if ! requester.isActive() {
            writer.WriteHeader(http.StatusForbidden)
            writer.Write(FORBIDDEN_MESSAGE)

            // TODO throttle/notify if too many weird requests.
            log.Printf("Denied inactive device request | device:%v", requester)
            return
        }


        vars := mux.Vars(request)
        username, userok   := vars["username"]
        deviceid, deviceok := vars["deviceid"]
        streamid, streamok := vars["streamid"]


        var user *User
        var device *Device
        var stream *Stream

        if userok {
            user, err = userdb.ReadUserByName(username)

            if err != nil {
                writer.WriteHeader(http.StatusBadRequest)
                writer.Write(BAD_REQUEST_MESSAGE)
                return
            }
        }

        if deviceok {
            devicei, err := strconv.Atoi(deviceid)

            if err != nil {
                badRequestResult.writeToHttp(writer)
                return
            }

            device,  err = userdb.ReadDeviceById(int64(devicei))

            if err != nil {
                writer.WriteHeader(http.StatusBadRequest)
                writer.Write(BAD_REQUEST_MESSAGE)
                return
            }
        }

        if streamok {
            streami, err := strconv.Atoi(streamid)

            if err != nil {
                writer.WriteHeader(http.StatusBadRequest)
                writer.Write(BAD_REQUEST_MESSAGE)
                return
            }

            stream, err = userdb.ReadStreamById(int64(streami))

            if err != nil {
                writer.WriteHeader(http.StatusBadRequest)
                writer.Write(BAD_REQUEST_MESSAGE)
                return
            }
        }

        if userOwnsReqeuster && !requester.isAdmin() {
            if user == nil || user.Id != requester.OwnerId {
                writer.WriteHeader(http.StatusBadRequest)
                writer.Write(BAD_REQUEST_MESSAGE)
                return
            }
        }

        if requesterIsDevice && !requester.isAdmin() {
            if device == nil || device.Id != requester.Id {
                writer.WriteHeader(http.StatusBadRequest)
                writer.Write(BAD_REQUEST_MESSAGE)
                return
            }
        }

        if requesterOwnsStream && !requester.isAdmin() {
            if stream == nil || stream.OwnerId != requester.Id {
                writer.WriteHeader(http.StatusBadRequest)
                writer.Write(BAD_REQUEST_MESSAGE)
                return
            }
        }

        // TODO check for upload limits

        resultcode, result := h(request, requester, user, device, stream)


        outputtype := vars["style"]

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
            errorGenericResult.writeToHttp(writer)
            return
        }

        writer.Header().Set("Content-Type", ct)
        writer.Header().Set("Content-Length", strconv.Itoa(len(val)))
        writer.WriteHeader(resultcode)

        writer.Write(val)

    })
}



func readAllUsers(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var result ReadUserResult
    result.Status = 200

    users, err := userdb.ReadAllUsers()

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    is_super := requestingDevice.isAdmin()

    for _, u := range users {
        result.Users = append(result.Users, u.ToClean())

        if is_super {
            result.Unsanitized = append(result.Unsanitized, *u)
        }
    }

    return http.StatusOK, result
}

func readUser(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var result ReadUserResult
    result.Status = 200

    if user == nil {
        return http.StatusBadRequest, badRequestResult
    }

    can_read_unsanitized := requestingDevice.isAdmin() || requestingDevice.IsOwnedBy(user)

    result.Users = append(result.Users, user.ToClean())

    if can_read_unsanitized {
        result.Unsanitized = append(result.Unsanitized, *user)
    }

    return http.StatusOK, result
}


// Requires a user struct with the name, email, password filled in. The password
// should be plaintext.
func createUser(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {

    var userUpload User

    if err := readBodyUnmarshalAndError(request, &userUpload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    id, err := userdb.CreateUser(userUpload.Name, userUpload.Email, userUpload.Password)

    switch {
        case err == ERR_EMAIL_EXISTS:
            return http.StatusInternalServerError, emailExistsResult
        case err == ERR_USERNAME_EXISTS:
            return http.StatusInternalServerError, usernameExistsResult
        case err != nil:
            log.Printf("Could not service create user request |err:%v", err)
            return http.StatusInternalServerError, errorGenericResult
        default:
            return http.StatusOK, NewCreateSuccessResult(id)
    }
}

func readBody(request *http.Request) ([]byte, error) {
    defer request.Body.Close()
    return ioutil.ReadAll(io.LimitReader(request.Body, 1048576))
}

func readBodyUnmarshalAndError(request *http.Request, tofill interface{}) (error) {
    body, err := readBody(request)
    if err != nil {
        return err
    }

    if err := json.Unmarshal(body, tofill); err != nil {
        return err
    }

    return nil
}


func updateUser(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var userUpload User

    if err := readBodyUnmarshalAndError( request, &userUpload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    if err := userdb.UpdateUser(&userUpload); err != nil {
        log.Printf("Could not service update user request|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, okGenericResult
}


func deleteUser(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var userUpload User

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


func readDevices(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    //var result ReadUserResult
    var result ReadDeviceResult
    result.Status = 200

    is_super := requestingDevice.isAdmin()
    devs, err := userdb.ReadDevicesForUserId(user.Id)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    for _, d := range devs {
        result.Devices = append(result.Devices, d.ToClean())

        if is_super {
            result.Unsanitized = append(result.Unsanitized, *d)
        }
    }

    return http.StatusOK, result
}

func readDevice(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    //var result ReadUserResult
    var result ReadDeviceResult
    result.Status = 200

    result.Devices = append(result.Devices, device.ToClean())

    if requestingDevice.isAdmin() || device.IsOwnedBy(user) {
        result.Unsanitized = append(result.Unsanitized, *device)
    }

    return http.StatusOK, result
}



func createDevice(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    //var result ReadUserResult
    var upload CleanDevice

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


func updateDevice(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var upload Device

    if err := readBodyUnmarshalAndError(request, &upload); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    if err := userdb.UpdateDevice(&upload); err != nil {
        log.Printf("Could not service update device request|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, okGenericResult
}


func deleteDevice(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    err := userdb.DeleteDevice(device.Id)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    } else {
        return http.StatusOK, okGenericResult
    }
}


func readStream(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var result ReadStreamResult

    is_super := requestingDevice.isAdmin()
    result.Streams = append(result.Streams, stream.ToClean())

    if is_super || requestingDevice.IsOwnedBy(user) {
        result.Unsanitized = append(result.Unsanitized, *stream)
    }

    return http.StatusOK, result
}

func readAllStreams(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var result ReadStreamResult

    is_super := requestingDevice.isAdmin()
    streams, err := userdb.ReadStreamsByDevice(device)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    for _, s := range streams {
        result.Streams = append(result.Streams, s.ToClean())

        if is_super || requestingDevice.IsOwnedBy(user) {
            result.Unsanitized = append(result.Unsanitized, *s)
        }
    }

    return http.StatusOK, result
}


func createStream(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {
    var result CleanStream

    if err := readBodyUnmarshalAndError(request, &result); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        return http.StatusInternalServerError, errorGenericResult
    }

    // todo return id
    id, err := userdb.CreateStream(result.Name, result.Schema_Json, result.Defaults_Json, device)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    }

    return http.StatusOK, NewCreateSuccessResult(id)
}


func deleteStream(request *http.Request, requestingDevice *Device, user *User, device *Device, stream *Stream) (int, interface{}) {

    err := userdb.DeleteStream(stream.Id)

    if err != nil {
        return http.StatusInternalServerError, errorGenericResult
    } else {
        return http.StatusOK, okGenericResult
    }
}




// Creates a subrouter available to
func GetSubrouter(subroutePrefix *mux.Router) {

    //    r := mux.NewRouter()
    if *ignoreBadApiKeys {
        subroutePrefix.HandleFunc("/firstrun/", firstRunHandler)
    }

    s := subroutePrefix.PathPrefix("/api/v1/{style}").Subrouter()


//apiAuth(h UserServiceHandler, requesterIsSuperdevice, userOwnsReqeuster, requesterIsDevice, requesterOwnsStream bool) http.HandlerFunc {

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
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/", apiAuth(updateDevice, false, true, true, false)).Methods("PUT")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/", apiAuth(deleteDevice, true, false, false, false)).Methods("DELETE")

    s.HandleFunc("/{username}/{deviceid:[0-9]+}/stream/", apiAuth(readAllStreams, false, true, false, false)).Methods("GET")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/stream/", apiAuth(createStream, false, true, true, false)).Methods("POST")

    s.HandleFunc("/{username}/{deviceid:[0-9]+}/{streamid:[0-9]+}/", apiAuth(readStream,   false, false, false, false)).Methods("GET")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/{streamid:[0-9]+}/", apiAuth(createStream, false, true, true, true)).Methods("POST")
    s.HandleFunc("/{username}/{deviceid:[0-9]+}/{streamid:[0-9]+}/", apiAuth(deleteStream, true , false, false, false)).Methods("DELETE")
}

func init() {
    var err error
    userdb, err = NewSqliteUserDatabase("production.sqlite")

    usr, err := userdb.ReadAllUsers()
    if len(usr) == 0 {
        log.Printf("No users found, auto ignoring bad api keys, all requests are superuser")
        *ignoreBadApiKeys = true
    }

    if err != nil {
        panic("Cannot open user database")
    }
}
