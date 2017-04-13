import React, { Component, PropTypes } from "react";
import { connect } from "react-redux";

import {
  Table,
  TableBody,
  TableHeader,
  TableHeaderColumn,
  TableRow,
  TableRowColumn
} from "material-ui/Table";

import Subheader from "material-ui/Subheader";

import { go } from "../actions";
import TimeDifference from "../components/TimeDifference";

import ObjectCard from "../components/ObjectCard";
import ObjectList from "../components/ObjectList";

import { objectFilter } from "../util";

class UserView extends Component {
  static propTypes = {
    user: PropTypes.shape({ name: PropTypes.string.isRequired }).isRequired,
    devarray: PropTypes.object.isRequired,
    state: PropTypes.shape({ expanded: PropTypes.bool.isRequired }).isRequired,
    onEditClick: PropTypes.func.isRequired,
    onExpandClick: PropTypes.func.isRequired,
    onAddClick: PropTypes.func.isRequired,
    onDeviceClick: PropTypes.func.isRequired,
    onHiddenClick: PropTypes.func.isRequired
  };

  render() {
    let user = this.props.user;
    let state = this.props.state;
    let description = user.description === undefined ? "" : user.description;
    let nickname = user.name;
    if (user.nickname !== undefined && user.nickname != "") {
      nickname = user.nickname;
    }

    return (
      <div>
        <ObjectCard
          expanded={state.expanded}
          onEditClick={this.props.onEditClick}
          onExpandClick={this.props.onExpandClick}
          style={{
            textAlign: "left"
          }}
          object={user}
          path={user.name}
        >
          <Table selectable={false}>
            <TableHeader
              enableSelectAll={false}
              displaySelectAll={false}
              adjustForCheckbox={false}
            >
              <TableRow>
                <TableHeaderColumn>Email</TableHeaderColumn>
                <TableHeaderColumn>Public</TableHeaderColumn>
                <TableHeaderColumn>Role</TableHeaderColumn>
                <TableHeaderColumn>Queried</TableHeaderColumn>
              </TableRow>
            </TableHeader>
            <TableBody displayRowCheckbox={false}>
              <TableRow>
                <TableRowColumn>{user.email}</TableRowColumn>
                <TableRowColumn>
                  {user.public ? "true" : "false"}
                </TableRowColumn>
                <TableRowColumn>{user.role}</TableRowColumn>
                <TableRowColumn>
                  <TimeDifference timestamp={user.timestamp} />
                </TableRowColumn>
              </TableRow>
            </TableBody>
          </Table>
        </ObjectCard>
        <Subheader
          style={{
            marginTop: "20px"
          }}
        >
          Devices
        </Subheader>
        <ObjectList
          showHidden={!state.hidden}
          onHiddenClick={this.props.onHiddenClick}
          style={{
            marginTop: "10px",
            textAlign: "left"
          }}
          objects={objectFilter(state.search.text, this.props.devarray)}
          addName="device"
          onAddClick={this.props.onAddClick}
          onSelect={this.props.onDeviceClick}
        />
      </div>
    );
  }
}

export default connect(undefined, (dispatch, props) => ({
  onEditClick: () => dispatch(go(props.user.name + "#edit")),
  onExpandClick: val =>
    dispatch({ type: "USER_VIEW_EXPANDED", name: props.user.name, value: val }),
  onAddClick: () => dispatch(go(props.user.name + "#create")),
  onDeviceClick: dev => dispatch(go(dev)),
  onHiddenClick: v =>
    dispatch({ type: "USER_VIEW_HIDDEN", name: props.user.name, value: v })
}))(UserView);
