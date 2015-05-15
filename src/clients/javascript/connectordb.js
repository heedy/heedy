
"use strict";

function ConnectorDB(user, password, url) {
    password = password | "";
    this.url = url | "https://connectordb.com/api/v1/";

    this.user = user;
    this.password = password | user;
    this.authHeader = "Basic " + btoa(user + ":" + password);

    if (password.length == 0) {
        this.authHeader = "Basic " + btoa(":" + user);
        this.user = "";
    }
}

ConnectorDB.prototype = {
    thisdevice: function () {
        return this.url;
    },

    // Internal mechanism for doing requests
    // path = /usr/dev/stream
    // reqtype = "GET" | "POST" | "PUT" | "DELETE"
    // object = undefined | posting object
    // returns a promise.
    _doRequest: function(path, reqtype, object) {
        var url = this.url + path;
        var user = this.user;
        var pass = this.password;

        return new Promise(function(resolve, reject) {
            var req = new XMLHttpRequest();

            // type, url, async, basicauth credentials
            req.open(reqtype, url, true, user, password);

            // normal response from server
            req.onload = function() {
                if (req.status == 200) {
                    resolve(req.response);
                }
                else {
                    reject(Error(req.statusText));
                }
            };

            // Handle network errors
            req.onerror = function() {
                reject(Error("Network Error"));
            };

            // Make the request
            if(object != undefined) {
                req.send(JSON.stringify(object));
            } else {
                req.send();
            }
        });
    },

    // Returns a connectordb path for the given user, dev, and stream
    // behavior is undefined if an item coming before a defined item is undefined
    // e.g. _getPath(undefined, "foo", "bar");
    _getPath: function(user, dev, stream) {
        var path = "/";
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
    }

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

    createStream: function(username, devicename, streamname) {
        var path = this._getPath(username, devicename, streamname)
        return this._doRequest(path, "POST");
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
    }
};
