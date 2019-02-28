/*
This is the textbox used to input a stream
*/
import React, { Component } from "react";
import PropTypes from "prop-types";
import TextField from "material-ui/TextField";

const StreamInput = ({ value, onChange }) =>
  <TextField
    hintText="user/device/stream"
    style={{ width: "100%" }}
    value={value}
    onChange={e => onChange(e.target.value)}
  />;

export default StreamInput;
