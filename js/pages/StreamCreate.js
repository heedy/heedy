import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {createCancel, createObject} from '../actions';

import ObjectCreate from '../components/ObjectCreate';

import RoleEditor from '../components/RoleEditor';

import PublicEditor from '../components/PublicEditor';

class StreamCreate extends Component {
    static propTypes = {
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        callbacks: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancel: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired
    }
    render() {
        let state = this.props.state;
        let callbacks = this.props.callbacks;
        return (< ObjectCreate type = "stream" state = {
            state
        }
        callbacks = {
            callbacks
        }
        parentPath = {
            this.props.user.name + "/" + this.props.device.name
        }
        onCancel = {
            this.props.onCancel
        }
        onSave = {
            this.props.onSave
        } > < /ObjectCreate >

        );
    }
}

export default connect((state) => ({
    roles: state.site.roles.device
}), (dispatch, props) => {
    let name = props.user.name+"/"+props.device.name;
 return {
            callbacks: {
                nameChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_NAME", name: name, value: txt}),
                nicknameChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_NICKNAME", name: name, value: txt}),
                descriptionChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_DESCRIPTION", name: name, value: txt})
            },
            onCancel: () => dispatch(createCancel("DEVICE", "STREAM", name)),
            onSave: () => dispatch(createObject("device", "stream", name, props.state))
        }})(StreamCreate);
