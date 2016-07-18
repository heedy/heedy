import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import Error from '../components/Error';
import {createCancel, createObject, go} from '../actions';

import ObjectCreate from '../components/ObjectCreate';

import DownlinkEditor from '../components/edit/DownlinkEditor';
import EphemeralEditor from '../components/edit/EphemeralEditor';
import DatatypeEditor from '../components/edit/DatatypeEditor';

import datatypes from '../datatypes/datatypes';

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
        let d = datatypes[this.props.datatype];
        if (d === undefined) {
            return (<Error err={{
                code: 500,
                ref: "",
                msg: "Datatype does not exist"
            }}/>);
        }
        return (
            <ObjectCreate type={d.name} header={d.create.description} required={d.create.required} advanced={this.props.onAdvanced} state={state} callbacks={callbacks} parentPath={this.props.user.name + "/" + this.props.device.name} onCancel={this.props.onCancel} onSave={this.props.onSave}>
                {d.create.optional}
            </ObjectCreate >

        );
    }
}

export default connect((state) => ({roles: state.site.roles.device}), (dispatch, props) => {
    let name = props.user.name + "/" + props.device.name;
    return {
        callbacks: {
            nameChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_NAME", name: name, value: txt}),
            nicknameChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_NICKNAME", name: name, value: txt}),
            descriptionChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_DESCRIPTION", name: name, value: txt})
        },
        onCancel: () => dispatch(createCancel("DEVICE", "STREAM", name)),
        onSave: () => dispatch(createObject("device", "stream", name, Object.assign({}, props.state, datatypes[props.datatype].create.default))),
        onAdvanced: () => {
            dispatch({
                type: "DEVICE_CREATESTREAM_SET",
                name: name,
                value: datatypes[props.datatype].create.default
            });
            dispatch(go(name + "#create"))
        }
    }
})(StreamCreate);
