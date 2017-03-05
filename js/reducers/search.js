/*
This file defines all of the manipulations of redux state necessary to implement search functionality. Each page displayed
has its own search context, which is saved independently
*/

import { location, getCurrentPath } from '../util';

const DefaultInitialState = {
    enabled: true,
    text: "",
    submitted: "", // The text that was submitted
    autocomplete: [],
    icon: "search",
    hint: "Search",
    error: ""
}

const InvalidState = {
    ...DefaultInitialState,
    enabled: false,
    icon: "error",
    hint: "Search not Available"
}

export const UserSearchInitialState = {
    ...DefaultInitialState,
    hint: "Search Devices"
};
export const DeviceSearchInitialState = {
    ...DefaultInitialState,
    hint: "Search Streams"
};
export const StreamSearchInitialState = {
    ...DefaultInitialState,
    icon: "keyboard_arrow_right",
    hint: "Filter & Transform Data"
};
export const IndexSearchInitialState = DefaultInitialState;

function basicSearchReducer(state, action, atype) {
    switch (atype) {
        case 'SET':
            return {
                ...state,
                text: action.value,
                error: ""
            };
        case 'SUBMIT':
            return {
                ...state,
                submitted: state.text
            };
        case 'SETSUBMIT':
            return {
                ...state,
                submitted: action.value
            };
        case 'SET_ERROR':
            return {
                ...state,
                error: action.value
            };
        case 'SET_STATE':
            return {
                ...state,
                ...action.value
            };
    }
    return state;
}
export function userSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("USER_VIEW_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}
export function deviceSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("DEVICE_VIEW_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}
export function streamSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("STREAM_VIEW_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}
export function indexSearchReducer(state, action) {
    let type = action.type;
    type = type.substring("PAGE_INDEX_SEARCH_".length, type.length);
    return basicSearchReducer(state, action, type);
}

// getSearchActionContext returns the necessary context to an action, including a prefix to
// use, to get search working with current page. Remember that each page has its own search context.
export function getSearchActionContext(action) {
    let actionPrefix = "INVALID_SEARCH_ACTION"; // This action won't be caught by any reducers

    let p = getCurrentPath();
    let path = p.split("/");
    if (p.length == 0) {
        // Later we can add the specific page hashes here
        actionPrefix = "PAGE_INDEX_SEARCH_";
    } else if (path.length == 1 && window.location.hash === "") {
        actionPrefix = "USER_VIEW_SEARCH_";
    } else if (path.length == 2 && window.location.hash === "") {
        actionPrefix = "DEVICE_VIEW_SEARCH_";
    } else if (path.length == 3 && window.location.hash === "") {
        actionPrefix = "STREAM_VIEW_SEARCH_";
    }

    action.name = p // Gives the specific device to use
    action.type = actionPrefix + action.type;
    return action;
}

// getSearchState returns the state of search given a location
export function getSearchState(state) {

    let p = getCurrentPath();
    let path = p.split("/");
    if (p.length == 0) {
        // Later we can add the specific page hashes here
        return state.pages.index.search;
    } else if (path.length == 1 && window.location.hash === "") {
        if (state.user[p] === undefined) {
            return UserSearchInitialState;
        }
        return state.user[p].view.search;
    } else if (path.length == 2 && window.location.hash === "") {
        if (state.device[p] === undefined) {
            return DeviceSearchInitialState;
        }
        return state.device[p].view.search;
    } else if (path.length == 3 && window.location.hash === "") {
        if (state.stream[p] === undefined) {
            return StreamSearchInitialState;
        }
        return state.stream[p].view.search;
    }
    return InvalidState;
}
