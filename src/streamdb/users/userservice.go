package users

import (
    "github.com/gorilla/mux"
    "net/http"
    "log"
    "fmt"
    "flag"
    )

var (
    UNAUTHORIZED_MESSAGE = []byte("The API key provided either doesn't exist our records or is disabled.")
    ignoreBadApiKeys = flag.Bool("ignoreBadApiKeys", false, "Ignores bad api keys and processes all requests as superuser.")
)



// Runs an authorization check on the api before calling the function
func apiAuth(h http.HandlerFunc, superdeviceRequired bool) http.HandlerFunc {
    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
        vars := mux.Vars(request)
        auth := vars["AuthKey"]
        device, err := ReadDeviceByApiKey(auth)


        // for development purposes
        if *ignoreBadApiKeys {
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

        h.ServeHTTP(writer, request)
    })
}

type ReadUserResult struct {
    Users []CleanUser
    Unsanitized []User
}



func readUser(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func createUser(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func updateUser(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


func deleteUser(writer http.ResponseWriter, request *http.Request) {
    //var result ReadUserResult

    // todo fill structs
    fmt.Fprintf(writer, "Hi there, I love %s!", request.URL.Path[1:])
}


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
    s.HandleFunc("/{AuthKey}/{username}/{devicename}/stream/", deleteStream).Methods("DELETE")

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
