import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {editCancel, deleteObject, saveObject} from '../actions';

import Dialog from 'material-ui/Dialog';
import FlatButton from 'material-ui/FlatButton';
import {RadioButton, RadioButtonGroup} from 'material-ui/RadioButton';
import TextField from 'material-ui/TextField';
import Checkbox from 'material-ui/Checkbox';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import Snackbar from 'material-ui/Snackbar';

import storage from '../storage';

import ObjectEdit from '../components/ObjectEdit';
import RoleEditor from '../components/edit/RoleEditor';
import PublicEditor from '../components/edit/PublicEditor';
import EmailEditor from '../components/edit/EmailEditor';

class UserEdit extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        state: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancel: PropTypes.func.isRequired,
        onDelete: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired,
        callbacks: PropTypes.object.isRequired
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
            <ObjectEdit object={user} path={user.name} state={edits} objectLabel={"user"} callbacks={this.props.callbacks} onDelete={this.props.onDelete} onCancel={this.props.onCancel} onSave={this.props.onSave}>

                <EmailEditor type="user" hintText="Email" floatingLabelText="Email" value={edits.email !== undefined
                    ? edits.email
                    : user.email} onChange={this.props.callbacks.emailChange}/><br/>
                <PublicEditor public={edits.public !== undefined
                    ? edits.public
                    : user.public} type="user" onChange={this.props.callbacks.publicChange}/>
                <RoleEditor roles={this.props.roles} role={edits.role !== undefined
                    ? edits.role
                    : user.role} type="user" onChange={this.props.callbacks.roleChange}/>
                <h3>Password</h3>
                <p>Change your user's password</p>
                <TextField hintText="Type New Password" type="password" style={{
                    marginTop: "-30px"
                }} value={edits.password !== undefined
                    ? edits.password
                    : ""} onChange={this.props.callbacks.passwordChange}/>
                <br/> {edits.password !== undefined
                    ? (<TextField hintText="Type New Password" type="password" floatingLabelText="Repeat New Password" value={edits.password2 !== undefined
                        ? edits.password2
                        : ""} onChange={this.props.callbacks.password2Change}/>)
                    : null}
                <br/>
            </ObjectEdit>
        );
    }
}

// It would be horrible to have all of these actions upstream - so we do them here.
export default connect((store) => ({roles: store.site.roles.user}), (dispatch, props) => ({
    callbacks: {
        nicknameChange: (e, txt) => dispatch({type: "USER_EDIT", name: props.user.name, value: {nickname: txt}}),
        descriptionChange: (e, txt) => dispatch({type: "USER_EDIT", name: props.user.name, value: {description:txt}}),
        passwordChange: (e, txt) => dispatch({type: "USER_EDIT", name: props.user.name, value: {password: txt}}),
        password2Change: (e, txt) => dispatch({type: "USER_EDIT", name: props.user.name, value: {password2: txt}}),
        roleChange: (e, role) => dispatch({type: "USER_EDIT", name: props.user.name, value: {role: role}}),
        publicChange: (e, val) => dispatch({type: "USER_EDIT", name: props.user.name, value: {public: val}}),
        emailChange: (e, val) => dispatch({type: "USER_EDIT", name: props.user.name, value: {email: val}}),
        iconChange: (e,val) => dispatch({type: "USER_EDIT", name: name, value: {icon:val}})
    },
    onCancel: () => dispatch(editCancel("USER", props.user.name)),
    onSave: () => dispatch(saveObject("user", props.user.name, props.user, props.state)),
    onDelete: () => dispatch(deleteObject("user", props.user.name))
}))(UserEdit);
