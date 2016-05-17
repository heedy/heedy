// Actions are things that can happen... To make it happen, run store.dispatch(action())
import {push, goBack} from 'react-router-redux'

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
