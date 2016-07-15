// This file holds the core implementation of obtaining data from ConnectorDB.
// Since users of the app will frequently be on mobile networks or otherwise want the app to be
// functional offline, data is extensively cached. In particular, every user, device, and stream that the user ever
// looks at is cached in an indexedDB database with a timestamp.
//
// Also, while technically users/devices/streams can be held in redux state, I decided to have storage with its own callback
// architecture, since the storage can have a lot of stuff cached - way too much for redux to handle.

// Note: Most components don't need to call this code - they usually access users/devices/streams through connectStorage.js

// TODO: This code should be cleaned up a bit...

import {ConnectorDB} from 'connectordb';

import localforage from 'localforage';
import 'localforage-startswith';
import 'localforage-setitems';

class Storage {
    constructor() {
        // Since we are logged in with cookie, the ConnectorDB js client does not need other authentication
        this.cdb = new ConnectorDB(undefined, undefined, SiteURL, true);

        // the cache where user/device/stream objects are stored.

        // TODO: Unfortunately, there need to be 2 instances of the storage, one for stream and one for user/device
        // because localForage does not have good support for SELECT WHERE type clauses. We can only use startsWith,
        // and that would return a user's devices, and all the devices' streams - when we only want a user's devices.
        // to avoid this, we simply split into two storage areas - one for users/devices and the other for streams
        this.store = localforage.createInstance({name: "cdb_cache"});
        this.streams = localforage.createInstance({name: "cdb_cache_stream"});

        // hotstore contains the user/device/streams that are currently being inserted into the store.
        // I ran into problems with store not containing objects that are being inserted in the background.
        // this fixes the issue by making objects available from a "hot" store until store contains them.
        this.hotstore = {}

        // Add callbacks that are run when a value is set.
        // The callbacks are indexed by an id, which allows removing them when not needed anymore
        this.callbacks = {};
    }

    // Just in case we want to log out - this clears all of the storage so that no data is left over
    clear() {
        console.log("Clearing storage...");
        return Promise.all([this.store.clear(), this.streams.clear()]);
    }

    // addContext adds the data returned with the page context when it is initially requested
    addContext(context) {

        // The data returned might be from cache. This means that it is OLD. The context's
        // timestamp will tell us whether this data is new or old.
        // As a rule: if it is more than 5 seconds old, it is considered old
        if (context.Timestamp * 1000 > Date.now() - 5000) {
            let inserter = {}

            if (context.ThisUser != null) {
                inserter[context.ThisUser.name] = context.ThisUser;
            }
            if (context.ThisDevice != null) {
                inserter[context.ThisUser.name + "/" + context.ThisDevice.name] = context.ThisDevice;
            }
            if (context.User != null) {
                inserter[context.User.name] = context.User;
            }
            if (context.Device != null) {
                inserter[context.User.name + "/" + context.Device.name] = context.Device;
            }
            if (context.Stream != null) {
                inserter[context.User.name + "/" + context.Device.name + "/" + context.Stream.name] = context.Stream;
            }

            this.setmany(inserter);
        } else {
            console.log("old context detected. This is a cached page.", context);
        }

    }

    // set sets the given object at the given path. It also adds a timestamp
    // parameter to the object so we know how old it is
    set(path, obj) {
        let newval = {
            ...obj,
            timestamp: Date.now()
        }
        this.hotstore[path] = newval;

        // Dealing with multiple storage locations
        let store = (path.split("/").length == 3
            ? this.streams
            : this.store);

        if (obj.ref !== undefined) {
            console.log("Removing from cache: " + path);
            store.removeItem(path).then(() => {
                // remove from hotstore
                delete this.hotstore[path];
            });
        } else {
            console.log("Updating cache: " + path, newval);
            store.setItem(path, newval).then(() => {
                // remove from hotstore
                delete this.hotstore[path];

            }).catch(function(err) {
                console.log(err);
            });
        }

        // Run all callbacks
        for (let id in this.callbacks) {
            this.callbacks[id](path, newval);
        }
        return;
    }

    setmany(obj) {
        console.log("Inserting multiple: ", obj);
        // The main annoyance here is having to deal with multiple storage locations - one for users/Devices
        // and the other for streams.

        if (obj.ref !== undefined)
            return;

        let streams = {};
        Object.keys(obj).forEach((key) => {
            obj[key].timestamp = Date.now();

            // We need to deal with storing in two places
            if (key.split("/").length == 3) {
                streams[key] = obj[key];
                delete obj[key];
            }

        });
        this.hotstore = Object.assign(this.hotstore, obj, streams);

        if (Object.keys(obj).length > 0) {
            this.store.setItems(obj).then(() => {
                Object.keys(obj).forEach((key) => {
                    delete this.hotstore[key];
                });
            }).catch(function(err) {
                console.log(err);
            });

            Object.keys(obj).forEach((key) => {
                for (let id in this.callbacks) {
                    this.callbacks[id](key, obj[key]);
                }
            });
        }
        if (Object.keys(streams).length > 0) {
            this.streams.setItems(streams).then(() => {
                Object.keys(streams).forEach((key) => {
                    delete this.hotstore[key];
                });
            }).catch(function(err) {
                console.log(err);
            });
            Object.keys(streams).forEach((key) => {
                for (let id in this.callbacks) {
                    this.callbacks[id](key, streams[key]);
                }
            });
        }

    }

    // query gets the most recent value of the given path directly from the ConnectorDB server.
    // this allows using new values, bypassing the cache completely
    query(path) {
        console.log("query: " + path);
        let p = path.split("/");
        switch (p.length) {
            case 1:
                var v = this.cdb.readUser(p[0]);
                break;
            case 2:
                var v = this.cdb.readDevice(p[0], p[1]);
                break;
            case 3:
                var v = this.cdb.readStream(p[0], p[1], p[2]);
                break;
        }
        return v.then((result) => {
            // If a result is returned, add to cache
            this.set(path, result);
            return result;
        });
    }
    // lsquery
    query_ls(path) {
        console.log("query_ls: " + path);
        let p = path.split("/");
        switch (p.length) {
            case 1:
                var v = this.cdb.listDevices(p[0]);
                break;
            case 2:
                var v = this.cdb.listStreams(p[0], p[1]);
                break;
        }
        return v.then((result) => {
            let res = {};
            // If the query was successful, add all of the devices to cache
            if (result.ref === undefined) {
                for (let i = 0; i < result.length; i++) {
                    res[path + "/" + result[i].name] = result[i];
                }
            }
            this.setmany(res);

            return res;
        });
    }

    qls(path) {
        // This is a combination of query and ls
        this.query(path);
        if (path.split("/").length <= 2) {
            this.query_ls(path);
        }
    }

    ls(path) {
        console.log("ls " + path);
        // for some reason, startsWith can't handle paths ending with '/', so to work around it, we query
        // all that start with the name, and then remove the ones that are not relevant
        // TODO: fix this...

        // Dealing with multiple storage locations
        let store = (path.split("/").length == 2
            ? this.streams
            : this.store);

        return store.startsWith(path).then((result) => {
            Object.keys(result).forEach((key) => {
                if (!key.startsWith(path + "/")) {
                    delete result[key];
                }
            });
            console.log("ls cache:", result);
            return result;
        });
    }

    addCallback(id, cb) {
        this.callbacks[id] = cb;
    }
    remCallback(id) {
        delete this.callbacks[id];
    }

    // get returns the given object if it is in the local storage
    get(path) {
        // Dealing with multiple storage locations
        let store = (path.split("/").length == 3
            ? this.streams
            : this.store);

        if (this.hotstore[path] !== undefined) {
            console.log("In hot cache: " + path);
            return Promise.resolve(this.hotstore[path]);
        }
        return store.getItem(path).then(function(value) {
            if (value != null) {
                console.log("Cache hit: " + path, value);
            } else {
                console.log("Cache miss: " + path);
            }
            return value;
        });
    }

    del(path) {
        console.log("delete: " + path);
        let p = path.split("/");
        switch (p.length) {
            case 1:
                var v = this.cdb.deleteUser(p[0]);
                break;
            case 2:
                var v = this.cdb.deleteDevice(p[0], p[1]);
                break;
            case 3:
                var v = this.cdb.deleteStream(p[0], p[1], p[2]);
                break;
        }
        return v.then((result) => {
            delete this.hotstore[path];
            this.store.removeItem(path);
            this.streams.removeItem(path);

            return result;
        });
    }
    update(path, structure) {
        console.log("update: " + path, structure);
        let p = path.split("/");
        switch (p.length) {
            case 0:
            case 1:
                var v = this.cdb.updateUser(p[0], structure);
                break;
            case 2:
                var v = this.cdb.updateDevice(p[0], p[1], structure);
                break;
            case 3:
                var v = this.cdb.updateStream(p[0], p[1], p[2], structure);
                break;
        }
        return v.then((result) => {
            if (result.ref === undefined) {
                this.set(path, result);
            }
            return result;
        });
    }

    create(path, structure) {
        console.log("create: " + path, structure);
        if (path == "") {
            var v = this.cdb.createUser(p[0], structure);
        } else {
            let p = path.split("/");
            switch (p.length) {
                case 1:
                    var v = this.cdb.createDevice(p[0], structure);
                    break;
                case 2:
                    var v = this.cdb.createStream(p[0], p[1], structure);
                    break;
            }
        }

        return v.then((result) => {
            if (result.ref === undefined) {
                this.set(path + "/" + structure.name, result);
            }
            return result;
        });
    }

    insert(user, device, stream, structure) {
        console.log("Inserting: " + user + "/" + device + "/" + stream + " data: " + JSON.stringify(structure));
        return this.cdb.insertStream(user, device, stream, structure);
    }

}
var storage = new Storage();

// storage is a global singleton
export default storage;
