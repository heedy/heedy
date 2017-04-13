import React, { Component, PropTypes } from "react";
import Checkbox from "material-ui/Checkbox";

class PublicEditor extends Component {
  static propTypes = {
    public: PropTypes.bool.isRequired,
    onChange: PropTypes.func.isRequired,
    type: PropTypes.string.isRequired
  };

  render() {
    return (
      <div>
        <h3>Public</h3>
        <p>
          Whether or not the {this.props.type + " "}
          can be accessed (viewed) by other users or devices.
        </p>
        <Checkbox
          label="Public"
          checked={this.props.public}
          onCheck={this.props.onChange}
        />
      </div>
    );
  }
}
export default PublicEditor;
