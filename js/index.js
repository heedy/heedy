import React from 'react';
import {render} from 'react-dom';

import {Provider} from 'react-redux'
import {createStore, dispatch} from 'redux'

import injectTapEventPlugin from 'react-tap-event-plugin';
// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

import App from './App'
import reducer from './reducer'
import {showPage} from './actions'

// runApp renders the app. It is assumed that the context is already set up correctly
function runApp(context, page) {
    let store = createStore(reducer);
    store.dispatch({type: 'LOAD_CONTEXT', value: context});
    store.dispatch(showPage(page));
    render((
        <Provider store={store}>
            <App/>
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

export function NotFound(context) {
    runApp(context, "404");
}

export function Index(context) {
    runApp(context, "main");
}
