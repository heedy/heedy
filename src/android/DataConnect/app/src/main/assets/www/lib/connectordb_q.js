function ConnectorDB(device, apikey, url) {
    url = url || "https://connectordb.com";

    this.url = url + "/api/v1/";

    this.device = device;

    this.authHeader = "Basic " + btoa(device + ":" + apikey);

}

ConnectorDB.prototype = {
    thisdevice: function () {
        return this.device;
    },

    // Internal mechanism for doing requests
    // path = /usr/dev/stream
    // reqtype = "GET" | "POST" | "PUT" | "DELETE"
    // object = undefined | posting object
    // returns a promise.
    _doRequest: function(path, reqtype, object) {
        var url = this.url + path;
        var user = this.device;
        var pass = this.apikey;
        var auth = this.authHeader

        var deferred = Q.defer();

        var req = new XMLHttpRequest();

        // type, url, async, basicauth credentials
        req.open(reqtype, url, true);

        req.setRequestHeader("Authorization",auth)

        // normal response from server
        req.onload = function() {
            if (req.status == 200) {
                try {
                    deferred.resolve(JSON.parse(req.response));
                } catch(err) {
                    deferred.resolve(req.response);
                }
            }
            else {
                deferred.reject(req);
            }
        };

        // Handle network errors
        req.onerror = function() {
            deferred.reject(null);
        };

        // Make the request
        if(object != undefined) {
            req.send(JSON.stringify(object));
        } else {
            req.send();
        }

        return deferred.promise;
    },

    // Returns a connectordb path for the given user, dev, and stream
    // behavior is undefined if an item coming before a defined item is undefined
    // e.g. _getPath(undefined, "foo", "bar");
    _getPath: function(user, dev, stream) {
        var path = "d/";
        if(user !== undefined) {
            path += user;
        }
        if( dev !== undefined ) {
            path += "/" + dev;
        }
        if( stream !== undefined ) {
            path += "/" + stream;
        }

        return path;
    },

    // Creates a new user
    createUser: function(username, email, password) {
        var path = this._getPath(username)
        return this._doRequest(path, "POST", {"Email":email, "Password":password});
    },

    // Reads an existing user
    readUser: function(username) {
        var path = this._getPath(username)
        return this._doRequest(path, "GET");
    },

    // Updates a user
    updateUser: function(username, structure) {
        var path = this._getPath(username)
        return this._doRequest(path, "PUT", structure);
    },

    // Deletes a given user
    deleteUser: function(username) {
        var path = this._getPath(username)
        return this._doRequest(path, "DELETE");
    },

    // Lists all the devices accessible form the user
    listDevices: function(username) {
        var path = this._getPath(username)+"?q=ls"
        return this._doRequest(path, "GET");
    },

    // Creates a device on the connectordb instance for the given user.
    createDevice: function(username, devicename) {
        var path = this._getPath(username, devicename)
        return this._doRequest(path, "POST");
    },

    // Reads a device from the connectordb server if it exists.
    readDevice: function(username, devicename) {
        var path = this._getPath(username, devicename)
        return this._doRequest(path, "GET");
    },

    // Updates a device on the connectordb server, if it does not exist
    // an error will be returned. Structure is the updated javascript object
    // that will be converted to JSON and sent.
    updateDevice: function(username, devicename, structure) {
        var path = this._getPath(username, devicename)
        return this._doRequest(path, "PUT", structure);
    },

    deleteDevice: function(username, devicename) {
        var path = this._getPath(username, devicename)
        return this._doRequest(path, "DELETE");
    },

    // Lists all the devices accessible form the user
    listStreams: function(username,devicename) {
        var path = this._getPath(username,devicename)+"?q=ls"
        return this._doRequest(path, "GET");
    },

    createStream: function(username, devicename, streamname,schema) {
        var path = this._getPath(username, devicename, streamname)
        return this._doRequest(path, "POST",schema);
    },

    readStream: function(username, devicename, streamname) {
        var path = this._getPath(username, devicename, streamname)
        return this._doRequest(path, "GET");
    },

    updateStream: function(username, devicename, streamname, structure) {
        var path = this._getPath(username, devicename, streamname)
        return this._doRequest(path, "PUT", structure);
    },

    deleteStream: function(username, devicename, streamname) {
        var path = this._getPath(username, devicename, streamname)
        return this._doRequest(path, "DELETE");
    },

    //Insert a single datapoint into the stream
    insertStream: function (username, devicename, streamname, data) {
        var datapoints = [{ t: (new Date).getTime() * 0.001, d: data }]
        var path = this._getPath(username, devicename, streamname)
        return this._doRequest(path, "UPDATE", datapoints);
    },

    //Get length of stream
    lengthStream: function (username, devicename, streamname) {
        var path = this._getPath(username, devicename, streamname) + "/length"
        return this._doRequest(path, "GET").then(function (result) { return parseInt(result); });
    },

    //Query by index range [i1,i2)
    indexStream: function (username, devicename, streamname,i1,i2) {
        var path = this._getPath(username, devicename, streamname) + "/data?i1=" + i1 + "&i2=" + i2;
        return this._doRequest(path, "GET");
    },

    //Query by time range [t1,t2) with a limited number of datapoints.
    //Current time is (new Date).getTime() * 0.001
    timeStream: function (username, devicename, streamname,t1,t2,limit) {
        limit = limit || 0;
        var path = this._getPath(username, devicename, streamname) + "/data?t1=" + t1 + "&t2=" + t2 + "&limit=" + limit;
        return this._doRequest(path, "GET");
    },
};
