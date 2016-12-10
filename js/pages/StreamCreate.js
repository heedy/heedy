/*
  StreamCreate is the page to show when creating a stream. The form to show is based upon the specific datatype that we're creating.
  This card renders the corresponding plugin from ../datatypes/creators.

*/
import React, { Component, PropTypes } from 'react';
import { connect } from 'react-redux';
import Error from '../components/Error';
import { createCancel, createObject, go } from '../actions';

import ObjectCreate from '../components/ObjectCreate';

import DownlinkEditor from '../components/edit/DownlinkEditor';
import EphemeralEditor from '../components/edit/EphemeralEditor';
import DatatypeEditor from '../components/edit/DatatypeEditor';

import { getCreator } from '../datatypes/datatypes';


const StreamCreateInitialState = {
    name: "",
    nickname: "",
    description: "",
    schema: '{}',
    downlink: false,
    ephemeral: false,
    datatype: ""
};

class StreamCreate extends Component {
    static propTypes = {
        datatype: PropTypes.string.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        callbacks: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancel: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired,
        setState: PropTypes.func.isRequired
    }
    render() {
        let state = Object.assign({}, StreamCreateInitialState, this.props.defaults, this.props.state);
        let callbacks = this.props.callbacks;
        let d = getCreator(this.props.datatype);

        return (
            <ObjectCreate type={d.name} header={d.description} required={d.required !== null
                ? (<d.required user={this.props.user} device={this.props.device} state={this.props.state} setState={this.props.setState} />)
                : null} state={state} callbacks={callbacks} parentPath={this.props.user.name + "/" + this.props.device.name} onCancel={this.props.onCancel} onSave={this.props.onSave}>
                {d.optional !== null
                    ? (<d.optional user={this.props.user} device={this.props.device} state={this.props.state} setState={this.props.setState} />)
                    : null}
            </ObjectCreate >

        );
    }
}

export default connect((state, props) => ({ roles: state.site.roles.device, defaults: getCreator(props.datatype).default }), (dispatch, props) => {
    let name = props.user.name + "/" + props.device.name;
    return {
        setState: (val) => dispatch({ type: "DEVICE_CREATESTREAM_SET", name: name, value: val }),
        callbacks: {
            nameChange: (e, txt) => dispatch({ type: "DEVICE_CREATESTREAM_NAME", name: name, value: txt }),
            nicknameChange: (e, txt) => dispatch({ type: "DEVICE_CREATESTREAM_NICKNAME", name: name, value: txt }),
            descriptionChange: (e, txt) => dispatch({ type: "DEVICE_CREATESTREAM_DESCRIPTION", name: name, value: txt }),
            iconChange: (e, val) => dispatch({ type: "DEVICE_CREATESTREAM_SET", name: name, value: { icon: val } })
        },
        onCancel: () => dispatch(createCancel("DEVICE", "STREAM", name)),
        onSave: () => dispatch(createObject("device", "stream", name, Object.assign({}, getCreator(props.datatype).default, props.state)))
    }
})(StreamCreate);
