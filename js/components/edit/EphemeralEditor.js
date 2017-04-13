import React, { Component, PropTypes } from "react";
import Checkbox from "material-ui/Checkbox";

class EphemeralEditor extends Component {
  static propTypes = {
    value: PropTypes.bool.isRequired,
    onChange: PropTypes.func.isRequired
  };

  render() {
    return (
      <div>
        <h3>Ephemeral</h3>
        <p>
          Ephemeral streams do not save data - inserts are only passed through ConnectorDB's messaging system
        </p>
        <Checkbox
          label="Ephemeral"
          checked={this.props.value}
          onCheck={this.props.onChange}
        />
      </div>
    );
  }
}
export default EphemeralEditor;
