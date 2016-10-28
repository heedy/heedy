/*
This file defines all of the manipulations of redux state necessary to implement search functionality. Each page displayed
has its own search context, which is saved independently
*/

import {location} from '../util';

const DefaultInitialState = {
    enabled: true,
    text: "",
    contextMenu: [],
    icon: "search",
    hint: "Search"
}

const InvalidState = {
    ...DefaultInitialState,
    enabled: false,
    icon: "error",
    hint: "Search not Available..."
}

export const UserSearchInitialState = DefaultInitialState;
export const DeviceSearchInitialState = DefaultInitialState;
export const StreamSearchInitialState = {
    ...DefaultInitialState,
    icon: "keyboard_arrow_right",
    hint: "Filter and Transform Data"
};
export const IndexSearchInitialState = DefaultInitialState;

function basicSearchReducer(state, action, atype) {
    console.log(atype);
    switch (atype) {
        case 'SET':
            return {
                ...state,
                text: action.value
            };
    }
    return state;
}
export function userSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("USER_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}
export function deviceSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("DEVICE_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}
export function streamSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("STREAM_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}
export function indexSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("PAGE_INDEX_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}

// Strips the beginning / and end / from the path
function getCurrentPath() {
    let p = location.pathname.substring(1, location.pathname.length);
    if (p.endsWith("/")) {
        p = p.substring(0, p.length - 1);
    }
    return p;
}

// getSearchActionContext returns the necessary context to an action, including a prefix to
// use, to get search working with current page. Remember that each page has its own search context.
export function getSearchActionContext(action) {
    let actionPrefix = "INVALID_SEARCH_ACTION"; // This action won't be caught by any reducers

    let p = getCurrentPath();
    let path = p.split("/");
    console.log("PATH", path);
    if (p.length == 0) {
        // Later we can add the specific page hashes here
        actionPrefix = "PAGE_INDEX_SEARCH_";
    } else if (path.length == 1 && location.hash === "") {
        actionPrefix = "USER_SEARCH_";
    } else if (path.length == 2 && location.hash === "") {
        actionPrefix = "DEVICE_SEARCH_";
    } else if (path.length == 3 && location.hash === "") {
        actionPrefix = "STREAM_SEARCH_";
    }

    action.name = p // Gives the specific device to use
    action.type = actionPrefix + action.type;
    console.log("GETCONTEXT", location, action);
    return action;
}

// getSearchState returns the state of search given a location
export function getSearchState_(state) {
    console.log("GETSTATE", location, state);

    let p = getCurrentPath();
    let path = p.split("/");
    console.log("PATH", p, path);
    if (p.length == 0) {
        // Later we can add the specific page hashes here
        return state.pages.index.search;
    } else if (path.length == 1 && location.hash === "") {
        if (state.user[p] === undefined) {
            return UserSearchInitialState;
        }
        return state.user[p].search;
    } else if (path.length == 2 && location.hash === "") {
        if (state.device[p] === undefined) {
            return DeviceSearchInitialState;
        }
        return state.user[p].search;
    } else if (path.length == 3 && location.hash === "") {
        if (state.stream[p] === undefined) {
            return StreamSearchInitialState;
        }
        return state.stream[p].search;
    }
    return InvalidState;
}

export function getSearchState(state) {
    let x = getSearchState_(state);
    console.log("SEarchSTATE", x);
    return x;
}
