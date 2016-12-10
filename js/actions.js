// Actions are things that can happen... To make it happen, run store.dispatch(action())

// TODO: This needs MAJOR cleanup. It needs to be split up into multiple files in an actions
// folder, and needs documentation on where stuff is used.

import { push, goBack } from 'react-router-redux'

import storage from './storage';
import { getCurrentPath } from './util';

import { StreamInputInitialState } from './reducers/stream';
import { getSearchActionContext } from './reducers/search';

// set the search bar text
export function setSearchText(text) {
    return getSearchActionContext({ type: 'SET', value: text });
}
// set the search bar submitted value
export function setSearchSubmit(text) {
    return getSearchActionContext({ type: 'SETSUBMIT', value: text });
}

// Allows to set values directly
export function setSearchState(val) {
    return getSearchActionContext({ type: 'SET_STATE', value: val });
}

// cancels an edit - and moves out of the edit screen
export function editCancel(type, path) {
    return (dispatch) => {
        dispatch({
            type: type + "_EDIT_CLEAR",
            name: path
        });
        dispatch(goBack());
    }
}

// cancels a create - and moves out of the create screen
export function createCancel(type, type2, path) {
    return (dispatch) => {
        dispatch({
            type: type + "_CREATE" + type2 + "_CLEAR",
            name: path
        });
        dispatch(goBack());
    }
}

export function go(loc) {
    return (dispatch) => {
        // When leaving user/device pages, we want to forget the search box contents.
        // But before we do that, let's make sure we get the CURRENT text
        let searchClear = setSearchText("");
        let path = getCurrentPath().split("/");

        dispatch(push("/" + loc));

        /*
        if (path.length == 1 || path.length == 2 ) {
            dispatch(searchClear);
        }
        */
        // For now, we clear all searches on leave page
        dispatch(searchClear);
    }
}

// Show a message in the snack bar
export function showMessage(msg) {
    return { type: 'SHOW_STATUS', value: msg };
}

export function deleteObject(type, path) {
    console.log("Deleting Object", type, path);
    return (dispatch) => {
        dispatch(showMessage("Deleting " + type + " '" + path + "'..."));
        // If the object is a stream, reset the state
        if (type === "stream") {
            dispatch({ type: "STREAM_CLEAR_STATE", name: path });
        }
        storage.del(path).then((result) => {
            if (result == "ok") {
                dispatch(showMessage("Deleted " + type + " '" + path + "'..."));

                // After the delete, go to the parent
                let p = path.split("/");
                switch (p.length) {
                    case 1:
                        dispatch(go(""));
                        break;
                    case 2:
                        dispatch(go(p[0]));
                        break;
                    case 3:
                        dispatch(go(p[0] + "/" + p[1]));
                        break;
                }
            } else {
                dispatch(showMessage(result.msg));
            }
        }).catch((err) => {
            console.log(err);
            dispatch(showMessage("Failed to delete " + type + " '" + path + "''"));
        });
    }
}

export function createObject(ftype, type, path, object) {
    console.log("Creating Object", ftype, type, path, object);
    return (dispatch) => {
        if (object.name == "") {
            dispatch(showMessage("Must give a name"));
            return;
        }
        if (object.name.toLowerCase() != object.name) {
            dispatch(showMessage("name must be lowercase"));
            return;
        }
        if (!/^[a-z0-9_]*$/.test(object.name)) {
            dispatch(showMessage("Name must not contain special characters or spaces"));
            return;
        }

        // If the object is a stream, reset the state
        if (type === "stream") {
            dispatch({ type: "STREAM_CLEAR_STATE", name: path });
        }

        storage.create(path, object).then((result) => {
            if (result.ref === undefined) {
                dispatch(showMessage("Created " + type + " '" + path + "/" + object.name + "'"));
                dispatch(createCancel(ftype.toUpperCase(), type.toUpperCase(), path));
                return;
            }
            dispatch(showMessage(result.msg));
        }).catch((err) => {
            console.log(err);
            dispatch(showMessage("Failed to create " + type + " '" + path + "/" + object.name + "''"));
        });
    }

}

export function saveObject(type, path, object, changes) {
    console.log("Saving Object", type, path, objec, changes);
    return (dispatch) => {
        dispatch(showMessage("Saving " + type + " '" + path + "'..."));

        // remove changes that are the same
        changes = Object.assign({}, changes);
        Object.keys(changes).forEach((key) => {
            if (object[key] !== undefined) {
                if (object[key] == changes[key]) {
                    delete changes[key];
                }
            }
        });

        // Password is only defined if this is a user, this code will be ignored if not user
        if (changes.password !== undefined) {
            if (changes.password != changes.password2) {
                dispatch(showMessage("Passwords do not match"));
                return;
            }
            if (changes.password == "") {
                delete changes.password;

            }
            delete changes.password2;
        }

        if (Object.keys(changes).length == 0) {
            dispatch(showMessage("Nothing changed"));
            dispatch(editCancel(type.toUpperCase(), path));
            return;
        }

        // If the object is a stream, reset the state
        if (type === "stream") {
            dispatch({ type: "STREAM_CLEAR_STATE", name: path });
        }

        // Finally, update the object
        storage.update(path, changes).then((result) => {
            if (result.ref === undefined) {
                dispatch(showMessage("Saved " + type + " '" + path + "'"));
                dispatch(editCancel(type.toUpperCase(), path));
                return;
            }
            dispatch(showMessage(result.msg));
        }).catch((err) => {
            console.log(err);
            dispatch(showMessage("Failed to save " + type + " '" + path + "''"));
        });

    };
}

export function dataInput(user, device, stream, timestamp, data, clearinput) {
    return (dispatch) => {
        storage.insert(user.name, device.name, stream.name, timestamp, data).then((result) => {
            if (result.ref === undefined) {
                if (clearinput != false) {
                    // Reset the input value
                    dispatch({
                        type: "STREAM_INPUT",
                        name: user.name + "/" + device.name + "/" + stream.name,
                        value: StreamInputInitialState
                    });
                }

                dispatch(showMessage("Inserted: " + JSON.stringify(data).substring(0, 15)));
                return;
            }
            dispatch(showMessage(result.msg));
        }).catch((err) => {
            console.log(err);
            dispatch(showMessage("Failed to insert:" + err.toString()));
        });
    };
}

export function query(user, device, stream, state) {
    let path = user.name + "/" + device.name + "/" + stream.name;
    console.log("Querying data for " + path, state);
    return (dispatch) => {
        if (state.bytime) {
            var d = storage.cdb.timeStream(user.name, device.name, stream.name, state.t1, state.t2, state.limit, state.transform);
        } else {
            var d = storage.cdb.indexStream(user.name, device.name, stream.name, state.i1, state.i2, state.transform);
        }
        d.then((result) => {
            if (result.ref !== undefined) {
                dispatch({ type: "STREAM_VIEW_ERROR", name: path, value: result });
                return;
            }
            dispatch({ type: "STREAM_VIEW_DATA", name: path, value: result });

        })/* .catch((err) => {
            console.log(err);
        });*/
    }
}
