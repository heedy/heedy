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

export var cache = storage;

import injectTapEventPlugin from 'react-tap-event-plugin';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

// Can always use some help!
console.log("Hi! You can follow along in the source code at https://github.com/connectordb/connectordb-frontend - and perhaps you can help out?");

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

// run renders the app. It is assumed that the context is already set up correctly
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
