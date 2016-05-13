// The Error page is displayed whenever there is an error returned from the app.
// The error information is available in globabl variables ErrorStatusCode and ErrorRefCode
// which are set up in error.html template

import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

class Error extends Component {

    render() {
        return (
            <div style={{
                textAlign: "center",
                paddingTop: 200,
                paddingBottom: 100
            }}>
                <h1>Oh no...</h1>
                <h2>There seems to have been an error</h2>
                <p>This object either doesn't exist, or you can't access it</p>
            </div>
        );
    }
}

export default Error;
