// Actions are things that can happen... To make it happen, run store.dispatch(action())
import {push, goBack} from 'react-router-redux'

import storage from './storage';

// set the search bar text
export function setSearchText(text) {
    return {type: 'SET_QUERY_TEXT', value: text};
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

export function go(loc) {
    return push("/" + loc);
}

// Show a message in the snack bar
export function showMessage(msg) {
    return {type: 'SHOW_STATUS', value: msg};
}

export function deleteObject(type, path) {
    return (dispatch) => {
        dispatch(showMessage("Deleting " + type + " '" + path + "'..."));
        storage.del(path).then((result) => {
            if (result == "ok")
                dispatch(showMessage("Deleted " + type + " '" + path + "'..."));
            else {
                dispatch(showMessage(result.msg));
            }

        }).catch((err) => {
            console.log(err);
            dispatch(showMessage("Failed to delete " + type + " '" + path + "''"));
        });
    }
}

export function saveObject(type, path, object, changes) {
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

    }
}
