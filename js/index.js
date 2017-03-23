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
import { render } from 'react-dom';

import { createStore, combineReducers, applyMiddleware, compose } from 'redux';
import { Provider } from 'react-redux';
import thunk from 'redux-thunk'

import createHistory from 'history/createBrowserHistory';
import { Router, Route } from 'react-router';
import { ConnectedRouter, routerReducer, routerMiddleware } from 'react-router-redux';

import { createLogger } from 'redux-logger';

import createSagaMiddleware from 'redux-saga';
import sagas from './sagas';

import { reducers } from './reducers/index';
import App from './App';
import { showPage } from './actions';
import storage from './storage';
import { setApp, getIPs } from './util';

// Register all of the available creators/inputs/views. All of ConnectorDB's visualizations are here.
import './datatypes/register';

export var cache = storage;

import injectTapEventPlugin from 'react-tap-event-plugin';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

// Can always use some help!
console.log("%cHi! You can follow along in the source code at https://github.com/connectordb/connectordb-frontend - pull requests are welcome!", "font-weight: bold;");

// Set up the ServiceWorker. The javascript is available in ../app/js/serviceworker.js
// http://www.html5rocks.com/en/tutorials/service-worker/introduction/

if ('serviceWorker' in navigator) {
    if (process.env.NODE_ENV == "debug") {
        console.log("%cRunning in debug mode", "font-weight: bold;");
        // If we are in debug mode, delete the ServiceWorkers that might be registered
        // https://stackoverflow.com/questions/33704791/how-do-i-uninstall-a-service-worker
        navigator.serviceWorker.getRegistrations().then(function (registrations) {
            for (let registration of registrations) {
                console.log("Unregistering ServiceWorker");
                registration.unregister();
            }
        });
    } else {
        navigator.serviceWorker.register('/serviceworker.js', { scope: "/" }).then(function (registration) {
            // Registration was successful
            console.log('ServiceWorker has scope: ', registration.scope);
        }).catch(function (err) {
            // registration failed :(
            console.log('ServiceWorker registration failed: ', err);
        });
    }

}

// Set up the Saga middleware, which will be used for dispatching actions.
const sagaMiddleware = createSagaMiddleware();

// The history
const history = createHistory();

// Set up the browser history redux middleware and the optional chrome dev tools extension for redux
// https://github.com/zalmoxisus/redux-devtools-extension/commit/6c146a2e16da79fefdc0e3e33f188d4ee6667341
let appMiddleware = applyMiddleware(thunk, routerMiddleware(history), sagaMiddleware, createLogger());
let finalCreateStore = compose(appMiddleware, window.devToolsExtension
    ? window.devToolsExtension()
    : f => f)(createStore);

export var store = finalCreateStore(combineReducers({
    ...reducers,
    routing: routerReducer
}));

sagaMiddleware.run(sagas);

// Makes the store available to outside this class
setApp(store);

// run renders the app. The context is passed in as json directly from ConnectorDB.
// The context has a timestamp, so the pages can be cached (have old context), and there
// shouldn't be a reason to worry
export function run(context) {
    // add the context to storage
    storage.addContext(context);
    // add context to state
    store.dispatch({ type: 'LOAD_CONTEXT', value: context });

    render((
        <Provider store={store}>
            <ConnectedRouter history={history}>
                <App />
            </ConnectedRouter>
        </Provider>
    ), document.getElementById('app'));
}

// We now asynchronously load PipeScript, which is used extensively for data analysis of downloaded data.
// While ConnectorDB has pipescript built-in, several visualizations perform further transforms of queried data.
// Instead of querying again, the transforms are done entirely client-side.
/*
Webpack has serious problems with the new version of PipeScript when building in production mode.
We therefore just load pipescript directly as an external library, from the html template.

TODO: Figure out wtf was wrong here...

require.ensure(["pipescript"], (p) => {
    console.log("PipeScript Loaded");
    store.dispatch({ type: 'PIPESCRIPT', value: require("pipescript") });
});
*/


// Finally, we correct the SiteURL if it is invalid.
// The issue stems from the fact that SiteURL is frequently localhost, 
// but we might want to connect an android app to ConnectorDB to sync. We need 
// a clever way to set up the URL if it is vague so that it can be accessed.

// We first parse the URL
// https://gist.github.com/jlong/2428561
let urlparser = document.createElement('a');
urlparser.href = SiteURL;

function isLocalhost(s) {
    return (s === "" || s === "localhost" || s === "127.0.0.1" || s === "::1");
}

if (isLocalhost(urlparser.hostname)) {
    if (!isLocalhost(window.location.hostname)) {
        // Use window.location value
        SiteURL = window.location.protocol + "//" + window.location.host;
    } else {
        // We don't have an alternative in current location. We use an WebRTC hack to get the local IP.
        // This is because if there is no setup, it means that the user is probably running ConnectorDB
        // desktop version - so we want the local network IP.

        getIPs(function (ip) {
            SiteURL = window.location.protocol + "//" + ip + ":" + window.location.port;
        });
    }
}
