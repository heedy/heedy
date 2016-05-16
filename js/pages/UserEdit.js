import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {editCancel} from '../actions';

import FlatButton from 'material-ui/FlatButton';
import {RadioButton, RadioButtonGroup} from 'material-ui/RadioButton';
import TextField from 'material-ui/TextField';
import Checkbox from 'material-ui/Checkbox';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import Avatar from 'material-ui/Avatar';

class UserEdit extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        state: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancelClick: PropTypes.func.isRequired,
        nicknameChange: PropTypes.func.isRequired,
        emailChange: PropTypes.func.isRequired,
        publicChange: PropTypes.func.isRequired,
        descriptionChange: PropTypes.func.isRequired,
        roleChange: PropTypes.func.isRequired,
        passwordChange: PropTypes.func.isRequired,
        password2Change: PropTypes.func.isRequired
    }

    render() {
        let user = this.props.user;
        let edits = this.props.state;
        let nickname = user.name;
        if (user.nickname !== undefined && user.nickname != "") {
            nickname = user.nickname;
        }
        if (edits.nickname !== undefined && edits.nickname != "") {
            nickname = edits.nickname;
        }
        return (
            <Card style={{
                textAlign: "left"
            }}>
                <CardHeader title={nickname} subtitle={user.name} avatar={< Avatar > U < /Avatar>}/>
                <CardText>
                    <TextField hintText="Nickname" floatingLabelText="Nickname" value={edits.nickname !== undefined
                        ? edits.nickname
                        : user.nickname} onChange={this.props.nicknameChange}/><br/>
                    <TextField hintText="Email" floatingLabelText="Email" value={edits.email !== undefined
                        ? edits.email
                        : user.email} onChange={this.props.emailChange}/><br/>
                    <h3>Public</h3>
                    <p>Whether or not the user can be accessed (viewed) by other users.</p>
                    <Checkbox label="Public" checked={edits.public !== undefined
                        ? edits.public
                        : user.public} onCheck={this.props.publicChange}/>
                    <h3>Description</h3>
                    <p>A user's description can be thought of as a README for the user.</p>
                    <TextField hintText="I am pretty awesome" floatingLabelText="Description" multiLine={true} fullWidth={true} value={edits.description !== undefined
                        ? edits.description
                        : user.description} style={{
                        marginTop: "-20px"
                    }} onChange={this.props.descriptionChange}/><br/>
                    <h3>Role</h3>
                    <p>A user's role determines the permissions given to operate upon ConnectorDB.</p>
                    <RadioButtonGroup name="role" valueSelected={edits.role !== undefined
                        ? edits.role
                        : user.role} onChange={this.props.roleChange}>
                        {Object.keys(this.props.roles).map((key) => (<RadioButton value={key} key={key} label={key + " - " + this.props.roles[key].description}/>))}
                    </RadioButtonGroup>
                    <h3>Password</h3>
                    <p>Change your user's password</p>
                    <TextField hintText="Type New Password" type="password" style={{
                        marginTop: "-30px"
                    }} value={edits.password !== undefined
                        ? edits.password
                        : ""} onChange={this.props.passwordChange}/>
                    <br/> {edits.password !== undefined
                        ? (<TextField hintText="Type New Password" type="password" floatingLabelText="Repeat New Password" value={edits.password2 !== undefined
                            ? edits.password2
                            : ""} onChange={this.props.password2Change}/>)
                        : null}
                    <br/>
                    <div style={{
                        paddingTop: "20px"
                    }}>
                        <FlatButton primary={true} label="Save"/>
                        <FlatButton label=" Cancel" onTouchTap={this.props.onCancelClick}/>
                        <FlatButton label="Delete" style={{
                            color: "red",
                            float: "right"
                        }}/>
                    </div>
                </CardText>
            </Card>
        );
    }
}

// It would be horrible to have all of these actions upstream - so we do them here.
export default connect((store) => ({roles: store.site.roles.user}), (dispatch, props) => ({
    nicknameChange: (e, txt) => dispatch({type: "USER_EDIT_NICKNAME", uname: props.user.name, value: txt}),
    descriptionChange: (e, txt) => dispatch({type: "USER_EDIT_DESCRIPTION", uname: props.user.name, value: txt}),
    passwordChange: (e, txt) => dispatch({type: "USER_EDIT_PASSWORD", uname: props.user.name, value: txt}),
    password2Change: (e, txt) => dispatch({type: "USER_EDIT_PASSWORD2", uname: props.user.name, value: txt}),
    roleChange: (e, role) => dispatch({type: "USER_EDIT_ROLE", uname: props.user.name, value: role}),
    publicChange: (e, val) => dispatch({type: "USER_EDIT_PUBLIC", uname: props.user.name, value: val}),
    emailChange: (e, val) => dispatch({type: "USER_EDIT_EMAIL", uname: props.user.name, value: val}),
    onCancelClick: () => dispatch(editCancel("USER", props.user.name))
}))(UserEdit);
