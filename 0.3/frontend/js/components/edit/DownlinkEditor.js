import React, { Component } from "react";
import PropTypes from "prop-types";
import Checkbox from "material-ui/Checkbox";

class DownlinkEditor extends Component {
  static propTypes = {
    value: PropTypes.bool.isRequired,
    onChange: PropTypes.func.isRequired
  };

  render() {
    return (
      <div>
        <h3>Downlink</h3>
        <p>
          Streams can be configured to have a parallel input stream which can be
          used to set goal states (such as turning lights on/off or setting
          thermostat temperature). A downlink stream has its normal output, but
          also allows intervention.
        </p>
        <Checkbox
          label="Downlink"
          checked={this.props.value}
          onCheck={this.props.onChange}
        />
      </div>
    );
  }
}
export default DownlinkEditor;
