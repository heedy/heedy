/*
  Main is the index page shown after initially logging in to ConnectorDB
*/

import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

import { go } from "../actions";

import MainToolbar from "../components/MainToolbar";
import FontIcon from "material-ui/FontIcon";
import IconButton from "material-ui/IconButton";

import Welcome from "../components/Welcome";
import DataInput from "../components/DataInput";

import { objectFilter } from "../util";

class Main extends Component {
  static propTypes = {
    user: PropTypes.shape({ name: PropTypes.string.isRequired }).isRequired,
    device: PropTypes.shape({ name: PropTypes.string.isRequired }).isRequired,
    streamarray: PropTypes.object.isRequired,
    state: PropTypes.object.isRequired,
    indexState: PropTypes.object.isRequired,

    onStreamClick: PropTypes.func.isRequired
  };

  render() {
    let state = this.props.state;
    let user = this.props.user;
    let device = this.props.device;
    let streams = this.props.streamarray;
    let indexState = this.props.indexState;
    return (
      <div
        style={{
          textAlign: "left"
        }}
      >
        <MainToolbar user={user} device={device} state={state} />
        {" "}
        {streams == null || Object.keys(streams).length == 0
          ? <Welcome />
          : <div
              style={{
                marginLeft: "-15px",
                marginRight: "-15px"
              }}
            >
              {Object.keys(
                objectFilter(indexState.search.text, streams)
              ).map(skey => {
                let s = streams[skey];
                let path = user.name + "/" + device.name + "/" + s.name;
                return (
                  <DataInput
                    key={s.name}
                    size={6}
                    thisUser={user}
                    thisDevice={device}
                    title={
                      s.nickname == ""
                        ? s.name.capitalizeFirstLetter()
                        : s.nickname
                    }
                    subtitle={path}
                    user={user}
                    device={device}
                    stream={s}
                    icons={[
                      <IconButton
                        key="showstream"
                        onTouchTap={() => this.props.onStreamClick(path)}
                        tooltip="view stream"
                      >
                        <FontIcon
                          className="material-icons"
                          color="rgba(0,0,0,0.8)"
                        >
                          list
                        </FontIcon>
                      </IconButton>
                    ]}
                  />
                );
              })}
            </div>}

      </div>
    );
  }
}
export default connect(
  state => ({ indexState: state.pages.index }),
  (dispatch, props) => ({
    onStreamClick: s => dispatch(go(s))
  })
)(Main);
