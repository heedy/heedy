// The Error page is displayed whenever there is an error returned from the app.
// The error information is available in globabl variables ErrorStatusCode and ErrorRefCode
// which are set up in error.html template

import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

class Error extends Component {
  static propTypes = {
    err: PropTypes.shape({
      code: PropTypes.number.isRequired,
      ref: PropTypes.string.isRequired,
      msg: PropTypes.string.isRequired
    }).isRequired
  };

  render() {
    return (
      <div
        style={{
          textAlign: "center",
          paddingTop: 200
        }}
      >
        <h1>{this.props.err.code}</h1>
        <h2>{this.props.err.msg}</h2>
        <p>{this.props.err.ref}</p>
      </div>
    );
  }
}

export default Error;
