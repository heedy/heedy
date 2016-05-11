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
                <h1>{ErrorStatusCode}</h1>
                <h2>Sorry, can't access this one!</h2>
                <p>{ErrorRefCode}</p>
            </div>
        );
    }
}

export default Error;
