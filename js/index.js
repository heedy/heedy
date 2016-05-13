import React from 'react';
import {render} from 'react-dom';

import {createStore, combineReducers, applyMiddleware, compose} from 'redux';
import {Provider} from 'react-redux';
import {Router, Route, browserHistory} from 'react-router';
import {syncHistoryWithStore, routerReducer, routerMiddleware} from 'react-router-redux';

import reducer from './reducer';
import App from './App';
import {showPage} from './actions';
import storage from './storage';

import injectTapEventPlugin from 'react-tap-event-plugin';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

// runApp renders the app. It is assumed that the context is already set up correctly
function runApp(context, page) {
    // add the context to storage
    storage.addContext(context);

    // Set up the browser history redux middleware and the optional chrome dev tools extension for redux
    // https://github.com/zalmoxisus/redux-devtools-extension/commit/6c146a2e16da79fefdc0e3e33f188d4ee6667341
    let browserMiddleware = applyMiddleware(routerMiddleware(browserHistory))
    let finalCreateStore = compose(browserMiddleware, window.devToolsExtension
        ? window.devToolsExtension()
        : f => f)(createStore);

    let store = finalCreateStore(combineReducers({app: reducer, routing: routerReducer}));

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
