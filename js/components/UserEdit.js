import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import FlatButton from 'material-ui/FlatButton';
import {RadioButton, RadioButtonGroup} from 'material-ui/RadioButton';
import TextField from 'material-ui/TextField';
import Toggle from 'material-ui/Toggle';

class UserEdit extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        onCancelClick: PropTypes.func.isRequired
    }

    render() {

        return (
            <div>
                <TextField hintText="Nickname" floatingLabelText="Nickname"/><br/>
                <TextField hintText="Email" floatingLabelText="Email"/><br/>
                <Toggle label="Public" style={{
                    maxWidth: 250
                }}/>
                <TextField hintText="I am pretty awesome" floatingLabelText="Description" multiLine={true} rows={2} fullWidth={true}/><br/>
                <h3>Role</h3>
                <p>A user's role determines the permissions given to operate upon ConnectorDB.</p>
                <RadioButtonGroup name="role">
                    <RadioButton value="light" label="user - can access own devices and public users/devices"/>
                    <RadioButton value="not_light" label="admin - has administrative access to the database"/>
                </RadioButtonGroup>

                <TextField hintText="Type New Password" type="password" floatingLabelText="Change Password"/><br/>
                <TextField hintText="Type New Password" type="password" floatingLabelText="Repeat New Password"/><br/>
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
            </div>
        );
    }
}
export default UserEdit;
