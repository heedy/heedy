import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {createCancel, createObject} from '../actions';

import ObjectCreate from '../components/ObjectCreate';

import RoleEditor from '../components/RoleEditor';

import PublicEditor from '../components/PublicEditor';
import EnabledEditor from '../components/EnabledEditor';
import VisibleEditor from '../components/VisibleEditor';

class DeviceCreate extends Component {
    static propTypes = {
        user: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        callbacks: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancel: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired
    }
    render() {
        let state = this.props.state;
        let callbacks = this.props.callbacks;
        return (
            <ObjectCreate type="device" state={state} callbacks={callbacks} parentPath={this.props.user.name} required= { <RoleEditor roles = { this.props.roles } role = { state.role } type = "device" onChange = { callbacks.roleChange } /> } onCancel={this.props.onCancel} onSave={this.props.onSave}>
                <PublicEditor type="device" public={state.public} onChange={callbacks.publicChange}/>
                <EnabledEditor type="device" value={state.enabled} onChange={callbacks.enabledChange}/>
                <VisibleEditor type="device" value={state.visible} onChange={callbacks.visibleChange}/>
            </ObjectCreate >

        );
    }
}

export default connect((state) => ({roles: state.site.roles.device}), (dispatch, props) => {
    let name = props.user.name;
    return {
        callbacks: {
            nameChange: (e, txt) => dispatch({type: "USER_CREATEDEVICE_NAME", name: name, value: txt}),
            nicknameChange: (e, txt) => dispatch({type: "USER_CREATEDEVICE_NICKNAME", name: name, value: txt}),
            descriptionChange: (e, txt) => dispatch({type: "USER_CREATEDEVICE_DESCRIPTION", name: name, value: txt}),
            roleChange: (e, role) => dispatch({type: "USER_CREATEDEVICE_ROLE", name: name, value: role}),
            publicChange: (e, val) => dispatch({type: "USER_CREATEDEVICE_PUBLIC", name: name, value: val}),
            enabledChange: (e, val) => dispatch({type: "USER_CREATEDEVICE_ENABLED", name: name, value: val}),
            visibleChange: (e, val) => dispatch({type: "USER_CREATEDEVICE_VISIBLE", name: name, value: val})
        },
        onCancel: () => dispatch(createCancel("USER", "DEVICE", name)),
        onSave: () => dispatch(createObject("user", "device", name, props.state))
    }
})(DeviceCreate);
