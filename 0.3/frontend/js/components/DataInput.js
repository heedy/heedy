import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

import ExpandableCard from "./ExpandableCard";
import AvatarIcon from "./AvatarIcon";

import { dataInput, showMessage } from "../actions";
import { getInput } from "../datatypes/datatypes";
import { getStreamState } from "../reducers/stream";

// Several properties in a view accept both a direct value OR a generator function that
// takes in the current state, and sets the view's value. This function extracts the correct
// value from these properties
function extractValue(value, context) {
  if (typeof value === "function") {
    return value(context);
  }
  return value;
}

class DataInput extends Component {
  static propTypes = {
    state: PropTypes.object.isRequired,
    user: PropTypes.object.isRequired,
    device: PropTypes.object.isRequired,
    thisUser: PropTypes.object.isRequired,
    thisDevice: PropTypes.object.isRequired,
    stream: PropTypes.object.isRequired,
    onSubmit: PropTypes.func.isRequired,
    setState: PropTypes.func.isRequired,
    showMessage: PropTypes.func.isRequired,
    icons: PropTypes.arrayOf(PropTypes.element),
    title: PropTypes.string,
    subtitle: PropTypes.string,
    showIcon: PropTypes.bool
  };

  static defaultProps = {
    title: "Insert Into Stream",
    subtitle: "",
    showIcon: false
  };

  touch() {
    if (this.props.touch !== undefined) {
      this.props.touch();
    }
  }
  render() {
    let user = this.props.user;
    let device = this.props.device;
    let stream = this.props.stream;
    let path = user.name + "/" + device.name + "/" + stream.name;

    let state = this.props.state;

    let schema = JSON.parse(stream.schema);

    // Based on the stream datatype, we get the component to display as an input. The default being
    // a standard form, but based on the datatype, it can be stars, or whatever is desired.
    let datatype = getInput(stream.datatype);

    let context = {
      path: path,
      schema: schema,
      showMessage: this.props.showMessage,

      // user/device/stream
      user: this.props.user,
      device: this.props.device,
      stream: this.props.stream,
      // currently logged in user/device
      thisUser: this.props.thisUser,
      thisDevice: this.props.thisDevice,
      // The input state, and setState
      state: state,
      setState: v => {
        this.props.setState({
          ...state,
          ...v
        });
      },
      // Insert the datapoint
      insert: this.props.onSubmit
    };

    let dropdown = null;
    if (datatype.dropdown !== undefined) {
      dropdown = <datatype.dropdown {...context} />;
    }

    // Finally, we append the icons sent in as props to our current icons
    let icons = extractValue(datatype.icons, context);
    if (icons === undefined) {
      icons = [];
    }
    if (this.props.icons !== undefined) {
      icons = icons.concat(this.props.icons);
    }

    return (
      <ExpandableCard
        avatar={
          this.props.showIcon
            ? <AvatarIcon name={stream.name} iconsrc={stream.icon} />
            : null
        }
        title={this.props.title}
        state={state}
        setState={this.props.setState}
        width={datatype.width}
        style={extractValue(datatype.style, context)}
        subtitle={this.props.subtitle}
        dropdown={dropdown}
        icons={icons}
      >
        <datatype.component {...context} />
      </ExpandableCard>
    );
  }
}

export default connect(
  (state, props) => ({
    state: getStreamState(
      props.user.name + "/" + props.device.name + "/" + props.stream.name,
      state
    ).input
  }),
  (dispatch, props) => ({
    onSubmit: (ts, val, cng) =>
      dispatch(dataInput(props.user, props.device, props.stream, ts, val, cng)),
    showMessage: val => dispatch(showMessage(val)),
    setState: v =>
      dispatch({
        type: "STREAM_INPUT",
        name:
          props.user.name + "/" + props.device.name + "/" + props.stream.name,
        value: v
      })
  })
)(DataInput);
