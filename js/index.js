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
import { Router, Route, browserHistory } from 'react-router';
import { syncHistoryWithStore, routerReducer, routerMiddleware } from 'react-router-redux';

import { reducers } from './reducers/index';
import App from './App';
import { showPage } from './actions';
import storage from './storage';
import { setApp } from './util';

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

// Makes the store available to outside this class
setApp(store);

// Set up the history through react-router-redux
let history = syncHistoryWithStore(browserHistory, store);

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
            <App history={history} />
        </Provider>
    ), document.getElementById('app'));
}

// We now asynchronously load PipeScript, which is used extensively for data analysis of downloaded data.
// While ConnectorDB has pipescript built-in, several visualizations perform further transforms of queried data.
// Instead of querying again, the transforms are done entirely client-side.
require.ensure(["pipescript"], (p) => {
    console.log("PipeScript Loaded");
    store.dispatch({ type: 'PIPESCRIPT', value: require("pipescript") });
});

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

        // https://github.com/diafygi/webrtc-ips
        function getIPs(callback) {
            var ip_dups = {};

            //compatibility for firefox and chrome
            var RTCPeerConnection = window.RTCPeerConnection
                || window.mozRTCPeerConnection
                || window.webkitRTCPeerConnection;
            var useWebKit = !!window.webkitRTCPeerConnection;

            //bypass naive webrtc blocking using an iframe
            if (!RTCPeerConnection) {
                //NOTE: you need to have an iframe in the page right above the script tag
                //
                //<iframe id="iframe" sandbox="allow-same-origin" style="display: none"></iframe>
                //<script>...getIPs called in here...
                //
                var win = iframe.contentWindow;
                RTCPeerConnection = win.RTCPeerConnection
                    || win.mozRTCPeerConnection
                    || win.webkitRTCPeerConnection;
                useWebKit = !!win.webkitRTCPeerConnection;
            }

            //minimal requirements for data connection
            var mediaConstraints = {
                optional: [{ RtpDataChannels: true }]
            };

            var servers = { iceServers: [{ urls: "stun:stun.services.mozilla.com" }] };

            //construct a new RTCPeerConnection
            var pc = new RTCPeerConnection(servers, mediaConstraints);

            function handleCandidate(candidate) {
                //match just the IP address
                var ip_regex = /([0-9]{1,3}(\.[0-9]{1,3}){3}|[a-f0-9]{1,4}(:[a-f0-9]{1,4}){7})/
                var ip_addr = ip_regex.exec(candidate)[1];

                //remove duplicates
                if (ip_dups[ip_addr] === undefined)
                    callback(ip_addr);

                ip_dups[ip_addr] = true;
            }

            //listen for candidate events
            pc.onicecandidate = function (ice) {

                //skip non-candidate events
                if (ice.candidate)
                    handleCandidate(ice.candidate.candidate);
            };

            //create a bogus data channel
            pc.createDataChannel("");

            //create an offer sdp
            pc.createOffer(function (result) {

                //trigger the stun server request
                pc.setLocalDescription(result, function () { }, function () { });

            }, function () { });

            //wait for a while to let everything done
            setTimeout(function () {
                //read candidate info from local description
                var lines = pc.localDescription.sdp.split('\n');

                lines.forEach(function (line) {
                    if (line.indexOf('a=candidate:') === 0)
                        handleCandidate(line);
                });
            }, 1000);
        }

        getIPs(function (ip) {
            SiteURL = window.location.protocol + "//" + ip + ":" + window.location.port;
        })
    }
}