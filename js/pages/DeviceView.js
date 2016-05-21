import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import Subheader from 'material-ui/Subheader';
import {
    Table,
    TableBody,
    TableHeader,
    TableHeaderColumn,
    TableRow,
    TableRowColumn
} from 'material-ui/Table';

import {go} from '../actions';
import TimeDifference from '../components/TimeDifference';
import ObjectCard from '../components/ObjectCard';
import ObjectList from '../components/ObjectList';

class DeviceView extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        device: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        streamarray: PropTypes.object.isRequired,
        state: PropTypes.shape({expanded: PropTypes.bool.isRequired}).isRequired,
        onEditClick: PropTypes.func.isRequired,
        onExpandClick: PropTypes.func.isRequired,
        onAddClick: PropTypes.func.isRequired,
        onStreamClick: PropTypes.func.isRequired
    }

    render() {
        let state = this.props.state;
        let user = this.props.user;
        let device = this.props.device;
        return (
            <div>
                <ObjectCard expanded={state.expanded} onEditClick={this.props.onEditClick} onExpandClick={this.props.onExpandClick} style={{
                    textAlign: "left"
                }} object={device} path={user.name + "/" + device.name}>
                    <Table selectable={false}>
                        <TableHeader enableSelectAll={false} displaySelectAll={false} adjustForCheckbox={false}>
                            <TableRow>
                                <TableHeaderColumn>Enabled</TableHeaderColumn>
                                <TableHeaderColumn>Public</TableHeaderColumn>
                                <TableHeaderColumn>Role</TableHeaderColumn>
                                <TableHeaderColumn>Queried</TableHeaderColumn>
                            </TableRow>
                        </TableHeader>
                        <TableBody displayRowCheckbox={false}>
                            <TableRow>
                                <TableRowColumn>{device.enabled
                                        ? "true"
                                        : "false"}</TableRowColumn>
                                <TableRowColumn>{device.public
                                        ? "true"
                                        : "false"}</TableRowColumn>
                                <TableRowColumn>{device.role == ""
                                        ? "none"
                                        : device.role}</TableRowColumn>
                                <TableRowColumn><TimeDifference timestamp={device.timestamp}/></TableRowColumn>
                            </TableRow>
                        </TableBody>
                    </Table >
                    {device.apikey !== undefined && device.apikey != ""
                        ? (
                            <div style={{
                                marginTop: "20px",
                                textAlign: "center",
                                color: "rgba(0, 0, 0, 0.541176)"
                            }}>
                                <h4>API Key</h4>
                                <p>{device.apikey}</p>
                            </div>
                        )
                        : null}

                </ObjectCard>
                <Subheader style={{
                    marginTop: "20px"
                }}>Streams</Subheader>
                <ObjectList style={{
                    marginTop: "10px",
                    textAlign: "left"
                }} objects={this.props.streamarray} addName="stream" onAddClick={this.props.onAddClick} onSelect={this.props.onStreamClick}/>
            </div>
        );
    }
}
export default connect(undefined, (dispatch, props) => ({
    onEditClick: () => dispatch(go(props.user.name + "/" + props.device.name + "#edit")),
    onExpandClick: (val) => dispatch({
        type: 'DEVICE_VIEW_EXPANDED',
        name: props.user.name + "/" + props.device.name,
        value: val
    }),
    onAddClick: () => dispatch(go(props.user.name + "/" + props.device.name + "#create")),
    onStreamClick: (s) => dispatch(go(s))
}))(DeviceView);
