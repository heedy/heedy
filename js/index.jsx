import React from 'react';
import {render} from 'react-dom';

import injectTapEventPlugin from 'react-tap-event-plugin';

import {siteRenderer} from './theme';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

class App extends React.Component {
    render() {
        return (
            <p>
                Hello React! This is it!
            </p>
        );
    }
}

/*
UserPage renders the page used to display a single user
*/
export function User(context) {
    console.log(context);
    render(
        <App/>, document.getElementById('app'));
}
/*
DevicePage renders the page used to display a single device
*/
export function Device(context) {
    console.log(context);
    render(
        <App/>, document.getElementById('app'));
}

/*
UserPage renders the page used to display a single stream
*/
export function Stream(context) {
    console.log(context);
    render(
        <App/>, document.getElementById('app'));
}

export function Error(context) {
    console.log(context);
    render(
        <App/>, document.getElementById('app'));
}

export function NotFound(context) {
    console.log(context);
    render(
        <App/>, document.getElementById('app'));
}

export function Index(context) {
    console.log(context);
    render(siteRenderer(App), document.getElementById('app'));
}
