/*
The starting point of the ConnectorDB frontend. This file does the following:
  - prepares the ServiceWorker (the serviceWorker is in ../app/serviceworker.js. It is
    not compiled into this bundle)
  - prepares redux, adds middlewares, and syncs it with the browser history (this is needed so that the router works
    correctly on the back button - since it is an SPA, we don't refresh the page when navigating)
  - registers all visualizations. This is done simply by importing datatypes/register.js
  - loads the app context into the store and into the state. The store holds cached users/devices/streams,
    and the state is the redux state. The context is passed in as json from ConnectorDB, and includes the
    current user, device, stream (if applicable), as well as the querying user and device.

After the setup, we are running in React. The main routing component that is invoked from this file is in App.js.
Look there to see how stuff is handled.
*/

import React from 'react';
import {render} from 'react-dom';

import {createStore, combineReducers, applyMiddleware, compose} from 'redux';
import {Provider} from 'react-redux';
import thunk from 'redux-thunk'
import {Router, Route, browserHistory} from 'react-router';
import {syncHistoryWithStore, routerReducer, routerMiddleware} from 'react-router-redux';

import {reducers} from './reducers/index';
import App from './App';
import {showPage} from './actions';
import storage from './storage';

// Register all of the available creators/inputs/views. All of ConnectorDB's visualizations are here.
import './datatypes/register';

export var cache = storage;

import injectTapEventPlugin from 'react-tap-event-plugin';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

// Can always use some help!
console.log("Hi! You can follow along in the source code at https://github.com/connectordb/connectordb-frontend - pull requests are welcome!");

// Set up the ServiceWorker. The javascript is available in ../app/js/serviceworker.js
// http://www.html5rocks.com/en/tutorials/service-worker/introduction/
if ('serviceWorker' in navigator && false) {
    navigator.serviceWorker.register('/serviceworker.js', {scope: "/"}).then(function(registration) {
        // Registration was successful
        console.log('ServiceWorker found: ', registration.scope);
    }).catch(function(err) {
        // registration failed :(
        console.log('ServiceWorker registration failed: ', err);
    });
}

// Set up the browser history redux middleware and the optional chrome dev tools extension for redux
// https://github.com/zalmoxisus/redux-devtools-extension/commit/6c146a2e16da79fefdc0e3e33f188d4ee6667341
let appMiddleware = applyMiddleware(thunk, routerMiddleware(browserHistory));
let finalCreateStore = compose(appMiddleware, window.devToolsExtension
    ? window.devToolsExtension()
    : f => f)(createStore);

export var store = finalCreateStore(combineReducers({
    ...reducers,
    routing: routerReducer
}));

// Set up the history through react-router-redux
let history = syncHistoryWithStore(browserHistory, store);

// run renders the app. The context is passed in as json directly from ConnectorDB.
// The context has a timestamp, so the pages can be cached (have old context), and there
// shouldn't be a reason to worry
export function run(context) {
    // add the context to storage
    storage.addContext(context);
    // add context to state
    store.dispatch({type: 'LOAD_CONTEXT', value: context});

    render((
        <Provider store={store}>
            <App history={history}/>
        </Provider>
    ), document.getElementById('app'));
}

// We now asynchronously load PipeScript, which is used extensively for data analysis of downloaded data.
require(["pipescript"], (p) => {
    console.log("PipeScript Loaded");
    store.dispatch({type: 'PIPESCRIPT', value: true});
})
