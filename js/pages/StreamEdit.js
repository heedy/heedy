import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

import { editCancel, go, deleteObject, saveObject } from "../actions";
import ObjectEdit from "../components/ObjectEdit";

import DownlinkEditor from "../components/edit/DownlinkEditor";
import EphemeralEditor from "../components/edit/EphemeralEditor";
import DatatypeEditor from "../components/edit/DatatypeEditor";
import SchemaEditor from "../components/edit/SchemaEditor";

class StreamEdit extends Component {
  static propTypes = {
    stream: PropTypes.object.isRequired,
    device: PropTypes.object.isRequired,
    user: PropTypes.object.isRequired,
    state: PropTypes.object.isRequired,
    callbacks: PropTypes.object.isRequired,
    onCancel: PropTypes.func.isRequired,
    onDelete: PropTypes.func.isRequired,
    onSave: PropTypes.func.isRequired
  };
  render() {
    let path =
      this.props.user.name +
      "/" +
      this.props.device.name +
      "/" +
      this.props.stream.name;
    let edits = this.props.state;
    let callbacks = this.props.callbacks;
    let stream = this.props.stream;
    return (
      <ObjectEdit
        object={this.props.stream}
        path={path}
        state={this.props.state}
        objectLabel={"stream"}
        callbacks={this.props.callbacks}
        onCancel={this.props.onCancel}
        onSave={this.props.onSave}
        onDelete={this.props.onDelete}
      >
        <SchemaEditor
          value={edits.schema !== undefined ? edits.schema : stream.schema}
          onChange={callbacks.schemaChange}
        />
        <DatatypeEditor
          value={
            edits.datatype !== undefined ? edits.datatype : stream.datatype
          }
          schema={stream.schema}
          onChange={callbacks.datatypeChange}
        />
        <DownlinkEditor
          value={
            edits.downlink !== undefined ? edits.downlink : stream.downlink
          }
          onChange={callbacks.downlinkChange}
        />
        <EphemeralEditor
          value={
            edits.ephemeral !== undefined ? edits.ephemeral : stream.ephemeral
          }
          onChange={callbacks.ephemeralChange}
        />

      </ObjectEdit>
    );
  }
}
export default connect(undefined, (dispatch, props) => {
  let name =
    props.user.name + "/" + props.device.name + "/" + props.stream.name;
  return {
    callbacks: {
      nicknameChange: (e, txt) =>
        dispatch({ type: "STREAM_EDIT_NICKNAME", name: name, value: txt }),
      descriptionChange: (e, txt) =>
        dispatch({ type: "STREAM_EDIT_DESCRIPTION", name: name, value: txt }),
      ephemeralChange: (e, txt) =>
        dispatch({ type: "STREAM_EDIT_EPHEMERAL", name: name, value: txt }),
      downlinkChange: (e, txt) =>
        dispatch({ type: "STREAM_EDIT_DOWNLINK", name: name, value: txt }),
      datatypeChange: (e, txt) =>
        dispatch({ type: "STREAM_EDIT_DATATYPE", name: name, value: txt }),
      iconChange: (e, val) =>
        dispatch({ type: "STREAM_EDIT", name: name, value: { icon: val } }),
      schemaChange: (e, val) =>
        dispatch({ type: "STREAM_EDIT", name: name, value: { schema: val } })
    },
    onCancel: () => dispatch(editCancel("STREAM", name)),
    onSave: () =>
      dispatch(saveObject("stream", name, props.stream, props.state)),
    onDelete: () => dispatch(deleteObject("stream", name))
  };
})(StreamEdit);
