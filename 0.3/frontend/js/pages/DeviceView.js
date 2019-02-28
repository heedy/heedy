import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

import Subheader from "material-ui/Subheader";
import {
  Table,
  TableBody,
  TableHeader,
  TableHeaderColumn,
  TableRow,
  TableRowColumn
} from "material-ui/Table";
import FlatButton from "material-ui/FlatButton";
import FontIcon from "material-ui/FontIcon";
import IconButton from "material-ui/IconButton";

import { go } from "../actions";
import TimeDifference from "../components/TimeDifference";
import ObjectCard from "../components/ObjectCard";
import ObjectList from "../components/ObjectList";

import { objectFilter } from "../util";

class DeviceView extends Component {
  static propTypes = {
    user: PropTypes.shape({ name: PropTypes.string.isRequired }).isRequired,
    device: PropTypes.shape({ name: PropTypes.string.isRequired }).isRequired,
    streamarray: PropTypes.object.isRequired,
    state: PropTypes.shape({ expanded: PropTypes.bool.isRequired }).isRequired,
    onEditClick: PropTypes.func.isRequired,
    onExpandClick: PropTypes.func.isRequired,
    onAddClick: PropTypes.func.isRequired,
    onStreamClick: PropTypes.func.isRequired
  };
  constructor(props) {
    super(props);
    this.state = {
      apikey: false
    };
  }

  render() {
    let state = this.props.state;
    let user = this.props.user;
    let device = this.props.device;

    return (
      <div>
        <ObjectCard
          expanded={state.expanded}
          onEditClick={this.props.onEditClick}
          onExpandClick={this.props.onExpandClick}
          style={{
            textAlign: "left"
          }}
          object={device}
          path={user.name + "/" + device.name}
        >
          <Table selectable={false}>
            <TableHeader
              enableSelectAll={false}
              displaySelectAll={false}
              adjustForCheckbox={false}
            >
              <TableRow>
                <TableHeaderColumn>Enabled</TableHeaderColumn>
                <TableHeaderColumn>Public</TableHeaderColumn>
                <TableHeaderColumn>Role</TableHeaderColumn>
                <TableHeaderColumn>Queried</TableHeaderColumn>
              </TableRow>
            </TableHeader>
            <TableBody displayRowCheckbox={false}>
              <TableRow>
                <TableRowColumn>
                  {device.enabled ? "true" : "false"}
                </TableRowColumn>
                <TableRowColumn>
                  {device.public ? "true" : "false"}
                </TableRowColumn>
                <TableRowColumn>
                  {device.role == "" ? "none" : device.role}
                </TableRowColumn>
                <TableRowColumn>
                  <TimeDifference timestamp={device.timestamp} />
                </TableRowColumn>
              </TableRow>
            </TableBody>
          </Table>
          {device.apikey !== undefined && device.apikey != ""
            ? <div
                style={{
                  marginTop: "20px",
                  textAlign: "center",
                  color: "rgba(0, 0, 0, 0.541176)"
                }}
              >
                {this.state.apikey
                  ? <div>
                      <h4>API Key</h4>
                      <p>{device.apikey}</p>
                      <IconButton
                        tooltip="Hide API Key"
                        onTouchTap={() => this.setState({ apikey: false })}
                      >
                        <FontIcon className="material-icons">
                          lock_outline
                        </FontIcon>
                      </IconButton>
                    </div>
                  : <FlatButton
                      label="Show API Key"
                      labelStyle={{ textTransform: "none" }}
                      onTouchTap={() => this.setState({ apikey: true })}
                      icon={
                        <FontIcon className="material-icons">
                          lock_open
                        </FontIcon>
                      }
                    />}
              </div>
            : null}

        </ObjectCard>
        <Subheader
          style={{
            marginTop: "20px"
          }}
        >
          Streams
        </Subheader>
        <ObjectList
          style={{
            marginTop: "10px",
            textAlign: "left"
          }}
          objects={objectFilter(state.search.text, this.props.streamarray)}
          addName="stream"
          onAddClick={this.props.onAddClick}
          onSelect={this.props.onStreamClick}
        />
      </div>
    );
  }
}
export default connect(undefined, (dispatch, props) => ({
  onEditClick: () =>
    dispatch(go(props.user.name + "/" + props.device.name + "#edit")),
  onExpandClick: val =>
    dispatch({
      type: "DEVICE_VIEW_EXPANDED",
      name: props.user.name + "/" + props.device.name,
      value: val
    }),
  onAddClick: () =>
    dispatch(go(props.user.name + "/" + props.device.name + "#create")),
  onStreamClick: s => dispatch(go(s))
}))(DeviceView);
