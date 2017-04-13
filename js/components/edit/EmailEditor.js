import React, { Component, PropTypes } from "react";
import TextField from "material-ui/TextField";

class EmailEditor extends Component {
  static propTypes = {
    value: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
    type: PropTypes.string.isRequired
  };

  render() {
    return (
      <div>
        <TextField
          hintText="Email"
          floatingLabelText="Email"
          value={this.props.value}
          onChange={this.props.onChange}
        />
        <br />
      </div>
    );
  }
}
export default EmailEditor;
