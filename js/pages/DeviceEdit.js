import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import Checkbox from 'material-ui/Checkbox';

import {editCancel, go, deleteObject, saveObject} from '../actions';
import ObjectEdit from '../components/ObjectEdit';
import RoleEditor from '../components/RoleEditor';

import PublicEditor from '../components/PublicEditor';

class DeviceEdit extends Component {
    static propTypes = {
        device: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        callbacks: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancel: PropTypes.func.isRequired,
        onDelete: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired
    }
    render() {
        let path = this.props.user.name + "/" + this.props.device.name;
        let edits = this.props.state;
        let device = this.props.device;
        return (
            <ObjectEdit object={this.props.device} path={path} state={this.props.state} objectLabel={"device"} callbacks={this.props.callbacks} onCancel={this.props.onCancel} onSave={this.props.onSave} onDelete={this.props.onDelete}>
                <PublicEditor type="device" public={edits.public !== undefined
                    ? edits.public
                    : device.public} onChange={this.props.callbacks.publicChange}/>

                <h3>API Key</h3>
                <p>You can check the box below to reset this device's API key</p>
                <Checkbox label="Reset API Key" checked={edits.apikey !== undefined} onCheck={this.props.callbacks.apikeyChange}/>

                <RoleEditor roles={this.props.roles} role={edits.role !== undefined
                    ? edits.role
                    : device.role} type="device" onChange={this.props.callbacks.roleChange}/>
            </ObjectEdit>
        );
    }
}
export default connect((state) => ({roles: state.site.roles.device}), (dispatch, props) => {
    let name = props.user.name + "/" + props.device.name;
    return {
        callbacks: {
            nicknameChange: (e, txt) => dispatch({type: "DEVICE_EDIT_NICKNAME", name: name, value: txt}),
            descriptionChange: (e, txt) => dispatch({type: "DEVICE_EDIT_DESCRIPTION", name: name, value: txt}),
            roleChange: (e, role) => dispatch({type: "DEVICE_EDIT_ROLE", name: name, value: role}),
            publicChange: (e, val) => dispatch({type: "DEVICE_EDIT_PUBLIC", name: name, value: val}),
            apikeyChange: (e, val) => dispatch({type: "DEVICE_EDIT_APIKEY", name: name, value: val})
        },
        onCancel: () => dispatch(editCancel("DEVICE", name)),
        onSave: () => dispatch(saveObject("device", name, props.device, props.state)),
        onDelete: () => dispatch(deleteObject("device", name))
    }
})(DeviceEdit);
