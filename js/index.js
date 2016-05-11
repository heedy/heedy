import React from 'react';
import {render} from 'react-dom';

import {createStore, combineReducers, applyMiddleware} from 'redux'
import {Provider} from 'react-redux'
import {Router, Route, browserHistory} from 'react-router'
import {syncHistoryWithStore, routerReducer, routerMiddleware} from 'react-router-redux'

import reducer from './reducer'

import App from './App'
import {showPage} from './actions'

import injectTapEventPlugin from 'react-tap-event-plugin';
// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

// runApp renders the app. It is assumed that the context is already set up correctly
function runApp(context, page) {
    let store = createStore(combineReducers({app: reducer, routing: routerReducer}), applyMiddleware(routerMiddleware(browserHistory)));

    // Set up the history through react-router-redux
    let history = syncHistoryWithStore(browserHistory, store);

    store.dispatch({type: 'LOAD_CONTEXT', value: context});

    render((
        <Provider store={store}>
            <App history={history}/>
        </Provider>
    ), document.getElementById('app'));
}

// UserPage renders the page used to display a single user
export function User(context) {
    runApp(context, context.User.name);
}

// DevicePage renders the page used to display a single device
export function Device(context) {
    runApp(context, context.User.name + "/" + context.Device.name);
}

// StreamPage renders the page used to display a single stream
export function Stream(context) {
    runApp(context, context.User.name + "/" + context.Device.name + "/" + context.Stream.name);
}

// Error is run when the app has an error
export function Error(context) {
    runApp(context, "404");
}

export function Index(context) {
    runApp(context, "main");
}
