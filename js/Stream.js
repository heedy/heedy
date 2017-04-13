/*
  This is the main navigator shown for streams. This component chooses the correct page to set based upon the data
  it is getting (shows loading page if the user/device/stream are not ready).

  The component also performs further routing based upon the hash. This is because react-router does not
  support both normal and hash-based routing at the same time.
  All child pages are located in ./pages. This component can be throught of as an extension to the main app routing
  done in App.js, with additional querying for the user/device/stream we want to view.

  It also queries the user/device/stream-specific state from redux, so further children can just use the state without worrying
  about which user/device/stream it belongs to.
*/

import React, { Component, PropTypes } from "react";
import { connect } from "react-redux";

import { getStreamState } from "./reducers/stream";
import connectStorage from "./connectStorage";

import Error from "./components/Error";
import Loading from "./components/Loading";

import StreamView from "./pages/StreamView";
import StreamEdit from "./pages/StreamEdit";

import { setTitle } from "./util";

function setStreamTitle(user, device, stream) {
  setTitle(
    user == null || device == null || stream == null
      ? ""
      : user.name + "/" + device.name + "/" + stream.name
  );
}

class Stream extends Component {
  static propTypes = {
    user: PropTypes.object,
    device: PropTypes.object,
    stream: PropTypes.object,
    error: PropTypes.object,
    location: PropTypes.object.isRequired,
    state: PropTypes.object
  };
  componentDidMount() {
    setStreamTitle(this.props.user, this.props.device, this.props.stream);
  }
  componentWillReceiveProps(newProps) {
    if (
      newProps.user !== this.props.user ||
      newProps.device !== this.props.device ||
      newProps.stream !== this.props.stream
    ) {
      setStreamTitle(newProps.user, newProps.device, newProps.stream);
    }
  }

  render() {
    if (this.props.error != null) {
      return <Error err={this.props.error} />;
    }
    if (
      this.props.user == null ||
      this.props.device == null ||
      this.props.stream == null
    ) {
      // Currently querying
      return <Loading />;
    }

    // React router does not allow using hash routing, so we route by hash here
    switch (this.props.location.hash) {
      case "#edit":
        return (
          <StreamEdit
            user={this.props.user}
            device={this.props.device}
            stream={this.props.stream}
            state={this.props.state.edit}
          />
        );
    }

    return (
      <StreamView
        user={this.props.user}
        device={this.props.device}
        stream={this.props.stream}
        state={this.props.state.view}
      />
    );
  }
}

export default connectStorage(
  connect((store, props) => ({
    state: getStreamState(
      props.user != null && props.device != null && props.stream != null
        ? props.user.name + "/" + props.device.name + "/" + props.stream.name
        : "",
      store
    )
  }))(Stream),
  false,
  false
);
