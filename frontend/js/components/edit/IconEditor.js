import React, { Component } from "react";
import PropTypes from "prop-types";
import TextField from "material-ui/TextField";

class IconEditor extends Component {
  static propTypes = {
    value: PropTypes.string,
    onChange: PropTypes.func.isRequired,
    type: PropTypes.string.isRequired
  };

  render() {
    let value = "";
    if (this.props.value !== undefined) {
      value = this.props.value;
    }
    return (
      <div>
        <h3>Icon</h3>
        <p>
          The icon can be a urlencoded image or an icon from the
          {" "}
          <a href="https://material.io/icons/">Material Design Icons</a>
          {" "}
          written in the form "material:icon_name".
        </p>
        <TextField
          floatingLabelText="Icon"
          multiLine={false}
          fullWidth={true}
          value={value}
          style={{
            marginTop: "-20px"
          }}
          onChange={this.props.onChange}
        />
        <br />
      </div>
    );
  }
}
export default IconEditor;
