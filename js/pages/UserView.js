import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {
    Table,
    TableBody,
    TableHeader,
    TableHeaderColumn,
    TableRow,
    TableRowColumn
} from 'material-ui/Table';
import {Card, CardText, CardHeader} from 'material-ui/Card';

import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import ReactMarkdown from 'react-markdown';

import Subheader from 'material-ui/Subheader';

import storage from '../storage';
import {go} from '../actions';
import TimeDifference from '../components/TimeDifference';
import AvatarIcon from '../components/AvatarIcon';

import ObjectList from '../components/ObjectList';

class UserView extends Component {

    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        devarray: PropTypes.object.isRequired,
        state: PropTypes.shape({expanded: PropTypes.bool.isRequired}).isRequired,
        onEditClick: PropTypes.func.isRequired,
        onExpandClick: PropTypes.func.isRequired
    }

    render() {

        let user = this.props.user;
        let state = this.props.state;
        let description = (user.description === undefined
            ? ""
            : user.description);
        let nickname = user.name;
        if (user.nickname !== undefined && user.nickname != "") {
            nickname = user.nickname;
        }

        return (
            <div>
                <Card style={{
                    textAlign: "left"
                }} onExpandChange={this.props.onExpandClick} expanded={state.expanded}>
                    <CardHeader title={nickname} subtitle={user.name} showExpandableButton={true} avatar={< AvatarIcon name = {
                        user.name
                    }
                    iconsrc = {
                        user.icon
                    } />}>
                        {state.expanded
                            ? (
                                <div style={{
                                    float: "right",
                                    marginRight: 35,
                                    marginTop: "-5px"
                                }}>
                                    <IconButton onTouchTap={() => this.props.onEditClick(true)} tooltip="edit">
                                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                            edit
                                        </FontIcon>
                                    </IconButton>
                                    <IconButton onTouchTap= { () => storage.query(user.name) } tooltip="reload">
                                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                            refresh
                                        </FontIcon>
                                    </IconButton>
                                </div>
                            )
                            : null}
                    </CardHeader>
                    <CardText expandable={true}>
                        {description == ""
                            ? (null)
                            : (
                                <div style={{
                                    color: "grey"
                                }}><ReactMarkdown escapeHtml={true} source={description}/></div>
                            )
}
                        <Table selectable={false}>
                            <TableHeader enableSelectAll={false} displaySelectAll={false} adjustForCheckbox={false}>
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
                                    <TableRowColumn>{user.public
                                            ? "true"
                                            : "false"}</TableRowColumn>
                                    <TableRowColumn>{user.role}</TableRowColumn>
                                    <TableRowColumn><TimeDifference timestamp={user.timestamp}/></TableRowColumn>
                                </TableRow>
                            </TableBody>
                        </Table >
                    </CardText>
                </Card>
                <Subheader style={{
                    marginTop: "20px"
                }}>Devices</Subheader>
                <ObjectList style={{
                    marginTop: "10px",
                    textAlign: "left"
                }} objects={this.props.devarray} addName="device"/>
            </div>
        );
    }
}

export default connect(undefined, (dispatch, props) => ({
    onEditClick: () => dispatch(go(props.user.name + "#edit")),
    onExpandClick: (val) => dispatch({type: 'USER_VIEW_EXPANDED', name: props.user.name, value: val})
}))(UserView);
