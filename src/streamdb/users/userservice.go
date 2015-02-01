package users

import (
    "github.com/gorilla/mux"
    "github.com/gorilla/context"
    "net/http"
    "log"
    "fmt"
    "flag"
    "encoding/json"
    "io"
    "io/ioutil"
    )

type key int

var (
    UNAUTHORIZED_MESSAGE = []byte("The API key provided either doesn't exist our records or is disabled.")
    ERROR_MESSAGE = []byte("An internal error occurred.")
    ignoreBadApiKeys = flag.Bool("ignoreBadApiKeys", false, "Ignores bad api keys and processes all requests as superuser.")
    errorGenericResult = GenericResult{http.StatusInternalServerError, "An internal error occurred"}
    okGenericResult = GenericResult{http.StatusOK, "Success"}
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


// Runs an authorization check on the api before calling the function
func apiAuth(h http.HandlerFunc, superdeviceRequired bool) http.HandlerFunc {
    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        vars := mux.Vars(request)
        auth := vars["AuthKey"]
        device, err := userdb.ReadDeviceByApiKey(auth)


        // for development purposes
        if *ignoreBadApiKeys {
            context.Set(request, REQUEST_DEVICE_IS_SUPERUSER, true)
            h.ServeHTTP(writer, request)
            return
        }

        // Check for devices that exist
        if device == nil || err != nil {
            writer.WriteHeader(http.StatusUnauthorized)
            writer.Write(UNAUTHORIZED_MESSAGE)

            if err != nil {
                log.Print(err)
            }

            return
        }

        // Check for superdevices.
        if superdeviceRequired && !device.isAdmin() {
            writer.WriteHeader(http.StatusUnauthorized)
            writer.Write(UNAUTHORIZED_MESSAGE)

            // TODO throttle/notify if too many weird requests.
            log.Printf("Denied superdevice request | device:%v", device)
            return

        }

        if ! device.isActive() {
            writer.WriteHeader(http.StatusUnauthorized)
            writer.Write(UNAUTHORIZED_MESSAGE)

            // TODO throttle/notify if too many weird requests.
            log.Printf("Denied inactive device request | device:%v", device)
            return
        }

        // TODO check for upload limits

        context.Set(request, REQUEST_DEVICE_IS_SUPERUSER, device.isAdmin())
        h.ServeHTTP(writer, request)
    })
}


func readUser(writer http.ResponseWriter, request *http.Request) {
    var result ReadUserResult
    result.Status = 200

    users, err := userdb.ReadAllUsers()

    if err != nil {
        log.Printf("Could not service read user request|err:%v", err)
        errorGenericResult.writeToHttp(writer)
        return
    }

    is_super := context.Get(request, REQUEST_DEVICE_IS_SUPERUSER)

    for _, u := range users {
        result.Users = append(result.Users, u.ToClean())

        if is_super.(bool) {
            result.Unsanitized = append(result.Unsanitized, *u)
        }
    }

    val, err := json.Marshal(result)

    if err != nil {
        log.Printf("Could not service read user request|err:%v", err)
        errorGenericResult.writeToHttp(writer)
        return
    }

    writer.WriteHeader(http.StatusOK)
    writer.Write(val)
}


func createUser(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}

func readBody(request *http.Request) ([]byte, error) {
    defer request.Body.Close()
    return ioutil.ReadAll(io.LimitReader(request.Body, 1048576))
}

func readBodyUnmarshalAndError(writer http.ResponseWriter, request *http.Request, tofill interface{}) (error) {
    body, err := readBody(request)
    if err != nil {
        log.Printf("Could not read the body|err:%v", err)
        errorGenericResult.writeToHttp(writer)
        return err
    }

    if err := json.Unmarshal(body, tofill); err != nil {
        log.Printf("Could not unmarshal|err:%v", err)
        errorGenericResult.writeToHttp(writer)
        return err
    }

    return nil
}


func updateUser(writer http.ResponseWriter, request *http.Request) {
    var userUpload User

    if err := readBodyUnmarshalAndError(writer, request, userUpload); err != nil {
        return // errors already handled
    }

    if err := userdb.UpdateUser(&userUpload); err != nil {
        log.Printf("Could not service update user request|err:%v", err)
        errorGenericResult.writeToHttp(writer)
        return
    }

    okGenericResult.writeToHttp(writer)
}


func deleteUser(writer http.ResponseWriter, request *http.Request) {
    var userUpload User

    if err := readBodyUnmarshalAndError(writer, request, userUpload); err != nil {
        return // errors already handled
    }

    if err := userdb.DeleteUser(userUpload.Id); err != nil {
        log.Printf("Could not service update user request|err:%v", err)
        errorGenericResult.writeToHttp(writer)
        return
    }

    okGenericResult.writeToHttp(writer)}


func readDevice(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func createDevice(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func updateDevice(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func deleteDevice(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func readStream(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func createStream(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func deleteStream(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}




// Creates a subrouter available to
func GetSubrouter(subroutePrefix *mux.Router) {

    //    r := mux.NewRouter()

    s := subroutePrefix.PathPrefix("/api/v1").Subrouter()

    //s.HandleFunc("/")

    s.HandleFunc("/{AuthKey}/user/", apiAuth(readUser, false)).Methods("GET")
    s.HandleFunc("/{AuthKey}/user/", apiAuth(createUser, true)).Methods("POST")
    s.HandleFunc("/{AuthKey}/user/", apiAuth(updateUser, true)).Methods("PUT")
    s.HandleFunc("/{AuthKey}/user/", apiAuth(deleteUser, true)).Methods("DELETE")
    // username exists?
    // validate user?

    s.HandleFunc("/{AuthKey}/{username}/device/", apiAuth(readDevice, false)).Methods("GET")
    s.HandleFunc("/{AuthKey}/{username}/device/", apiAuth(createDevice, true)).Methods("POST")
    s.HandleFunc("/{AuthKey}/{username}/device/", apiAuth(updateDevice, true)).Methods("PUT")
    s.HandleFunc("/{AuthKey}/{username}/device/", apiAuth(deleteDevice, true)).Methods("DELETE")

    s.HandleFunc("/{AuthKey}/{username}/{devicename}/stream/", apiAuth(readStream, false)).Methods("GET")
    s.HandleFunc("/{AuthKey}/{username}/{devicename}/stream/", apiAuth(createStream, false)).Methods("POST")
    // no puts, stream data is immutable.
    s.HandleFunc("/{AuthKey}/{username}/{devicename}/stream/", apiAuth(deleteStream, true)).Methods("DELETE")

    /**
    /api/v1/{AuthKey}/{username}/device/
    GET - list all devices
    POST - create a device
    PUT  - update a device
    DELETE - removes a device -- superdevice only

    /api/v1/{AuthKey}/{username}/{devicename}/stream/
    GET - list all streams by the device
    POST - create a stream
    PUT - update a stream
    DELETE - removes a stream -- superdevice only

    /api/v1/{AuthKey}/{username}/{devicename}/{streamname}/
    GET - list all data in a stream
    ?starttime=TIMESTAMP&endtime=TIMESTAMP filtering
    POST - push data to a stream
    PUT - NOT AVAILABLE, DATA IS IMMUTABLE
    DELETE - removes a stream -- superdevice only
    **/
}

func init() {
    var err error
    userdb, err = NewSqliteUserDatabase("production.sqlite")

    if err != nil {
        panic("Cannot open user database")
    }
}
