import React, { Component } from "react";
import PropTypes from "prop-types";
import { RadioButton, RadioButtonGroup } from "material-ui/RadioButton";

class RoleEditor extends Component {
  static propTypes = {
    roles: PropTypes.object.isRequired,
    role: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
    type: PropTypes.string.isRequired
  };

  render() {
    return (
      <div>
        <h3>Role</h3>
        <p>
          A
          {" "}
          {this.props.type}
          's role determines the permissions given to operate upon ConnectorDB.
        </p>
        <RadioButtonGroup
          name="role"
          valueSelected={this.props.role == "" ? "none" : this.props.role}
          onChange={this.props.onChange}
        >
          {Object.keys(this.props.roles).map(key =>
            <RadioButton
              value={key}
              key={key}
              label={key + " - " + this.props.roles[key].description}
            />
          )}
        </RadioButtonGroup>
      </div>
    );
  }
}
export default RoleEditor;
