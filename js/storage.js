// This file holds the core implementation of obtaining data from ConnectorDB.
// Since users of the app will frequently be on mobile networks or otherwise want the app to be
// functional offline, data is extensively cached. In particular, every user, device, and stream that the user ever
// looks at is cached in an indexedDB database with a timestamp.
//
// Also, while technically users/devices/streams can be held in redux state, I decided to have storage with its own callback
// architecture, since the storage can have a lot of stuff cached - way too much for redux to handle.

import {ConnectorDB} from 'connectordb';

import localforage from 'localforage';

class Storage {
    constructor() {
        // Since we are logged in with cookie, the ConnectorDB js client does not need other authentication
        this.cdb = new ConnectorDB(undefined, undefined, SiteURL, true);

        // the cache where user/device/stream objects are stored
        this.store = localforage.createInstance({name: "cdb_cache"});
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
        return this.store.clear().then(() => {
            return this.queryqueue.clear()
        });
    }

    // addContext adds the data returned with the page context when it is initially requested
    addContext(context) {
        if (context.ThisUser != null) {
            this.set(context.ThisUser.name, context.ThisUser);
        }
        if (context.ThisDevice != null) {
            this.set(context.ThisUser.name + "/" + context.ThisDevice.name, context.ThisDevice);
        }
        if (context.User != null && context.ThisUser.name != context.User.name) {
            this.set(context.User.name, context.User);
        }
        if (context.Device != null && !(context.ThisUser.name == context.User.name && context.ThisDevice.name == context.Device.name)) {
            this.set(context.User.name + "/" + context.Device.name, context.Device);
        }
        if (context.Stream != null) {
            this.set(context.User.name + "/" + context.Device.name + "/" + context.Stream.name, context.Stream);
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

        if (obj.ref !== undefined) {
            console.log("Removing from cache: " + path);
            this.store.removeItem(path).then(() => {
                // remove from hotstore
                delete this.hotstore[path];

                // Run all callbacks
                for (let id in this.callbacks) {
                    this.callbacks[id](path, newval);
                }
            });
            return;
        }

        console.log("Updating cache: " + path, newval);
        this.store.setItem(path, newval).then(() => {

            // remove from hotstore
            delete this.hotstore[path];

            // Run all callbacks
            for (let id in this.callbacks) {
                this.callbacks[id](path, newval);
            }
        }).catch(function(err) {
            console.log(err);
        });
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

    addCallback(id, cb) {
        this.callbacks[id] = cb;
    }
    remCallback(id) {
        delete this.callbacks[id];
    }

    // get returns the given object if it is in the local storage
    get(path) {
        if (this.hotstore[path] !== undefined) {
            console.log("In hot cache: " + path);
            return Promise.resolve(this.hotstore[path]);
        }
        return this.store.getItem(path).then(function(value) {
            if (value != null) {
                console.log("Cache hit: " + path, value);
            } else {
                console.log("Cache miss: " + path);
            }
            return value;
        });
    }

}
var storage = new Storage();

// storage is a global singleton
export default storage;
