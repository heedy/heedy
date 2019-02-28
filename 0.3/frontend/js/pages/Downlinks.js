import React, { Component } from "react";
import { connect } from "react-redux";
import { bindActionCreators } from "redux";

import { Card, CardText, CardHeader } from "material-ui/Card";
import FontIcon from "material-ui/FontIcon";
import IconButton from "material-ui/IconButton";
import Loading from "../components/Loading";
import DataInput from "../components/DataInput";

import * as Actions from "../actions/downlinks";
import { go } from "../actions";

import { arrayFilter } from "../util";

const Welcome = () => (
  <Card
    style={{
      marginTop: "20px"
    }}
  >
    <CardHeader
      title={"Downlinks"}
      subtitle={"Control your devices through ConnectorDB"}
    />
    <CardText>
      <p>
        It looks like you don't have any downlinks set up yet. If you sync devices such as your lights or your thermostat to ConnectorDB, you will be able to control them here.
      </p>
      <p>
        Downlink streams allow external input, which is immediately sent to the relevant device. Once acknowledged by the device, the input is added to that stream's data.
      </p>

    </CardText>
  </Card>
);

const DownlinkInput = (state, actions, downlink, path, go) => (
  <DataInput
    showIcon={true}
    key={path}
    size={6}
    thisUser={state.site.thisUser}
    thisDevice={state.site.thisDevice}
    title={
      downlink.nickname == ""
        ? downlink.name.capitalizeFirstLetter()
        : downlink.nickname
    }
    subtitle={path}
    user={state.site.thisUser}
    device={{ name: downlink.device, downlink: true }}
    stream={downlink}
    icons={[
      <IconButton
        key="showstream"
        onTouchTap={() => go(path)}
        tooltip="view stream"
      >
        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
          list
        </FontIcon>
      </IconButton>
    ]}
  />
);

const Render = ({ state, appstate, actions, go }) => (
  <div
    style={{
      textAlign: "left"
    }}
  >
    {!state.loaded
      ? <Loading />
      : state.downlinks.length == 0
          ? <Welcome />
          : <div
              style={{
                marginLeft: "-15px",
                marginRight: "-15px"
              }}
            >
              {arrayFilter(state.search.text, state.downlinks).map(d =>
                DownlinkInput(
                  appstate,
                  actions,
                  d,
                  appstate.site.thisUser.name + "/" + d.device + "/" + d.name,
                  go
                )
              )}
            </div>}
  </div>
);

export default connect(
  state => ({ state: state.pages.downlinks, appstate: state }),
  dispatch => ({
    actions: bindActionCreators(Actions, dispatch),
    go: v => dispatch(go(v))
  })
)(Render);
