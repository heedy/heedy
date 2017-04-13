import React, { Component, PropTypes } from "react";
import { addCreator } from "../datatypes";

import SchemaEditor from "../../components/edit/SchemaEditor";
import DownlinkEditor from "../../components/edit/DownlinkEditor";
import EphemeralEditor from "../../components/edit/EphemeralEditor";
import DatatypeEditor from "../../components/edit/DatatypeEditor";

class DefaultRequired extends Component {
  static propTypes = {
    user: PropTypes.object.isRequired,
    device: PropTypes.object.isRequired,
    state: PropTypes.object.isRequired,
    setState: PropTypes.func.isRequired
  };

  render() {
    let state = this.props.state;
    return (
      <SchemaEditor
        value={state.schema}
        onChange={(e, schema) => this.props.setState({ schema: schema })}
      />
    );
  }
}

class DefaultOptional extends Component {
  static propTypes = {
    user: PropTypes.object.isRequired,
    device: PropTypes.object.isRequired,
    state: PropTypes.object.isRequired,
    setState: PropTypes.func.isRequired
  };

  render() {
    let state = this.props.state;
    let set = this.props.setState;
    return (
      <div>
        <DatatypeEditor
          value={state.datatype}
          schema={state.schema}
          onChange={(e, txt) => set({ datatype: txt })}
        />
        <DownlinkEditor
          value={state.downlink}
          onChange={(e, txt) => set({ downlink: txt })}
        />
        <EphemeralEditor
          value={state.ephemeral}
          onChange={(e, txt) => set({ ephemeral: txt })}
        />
      </div>
    );
  }
}

// Empty string registers as default
addCreator("", {
  name: "stream",
  required: DefaultRequired,
  optional: DefaultOptional,
  description: "You can create any type of stream here. If you want to create a specific type of stream, choose its icon from the Insert (main) page.",
  default: {}
});
