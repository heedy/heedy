import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {getDeviceState} from './reducers/device';
import connectStorage from './connectStorage';

import Error from './components/Error';
import Loading from './components/Loading';

import DeviceView from './pages/DeviceView';
import DeviceEdit from './pages/DeviceEdit';
import StreamCreate from './pages/StreamCreate';

import {setTitle} from './util';

function setDeviceTitle(user, device) {
    setTitle(user == null || device == null
        ? ""
        : user.name + "/" + device.name);
}

class Device extends Component {
    static propTypes = {
        user: PropTypes.object,
        device: PropTypes.object,
        streamarray: PropTypes.object,
        error: PropTypes.object,
        location: PropTypes.object.isRequired,
        state: PropTypes.object
    };
    componentDidMount() {
        setDeviceTitle(this.props.user, this.props.device);
    }
    componentWillReceiveProps(newProps) {
        if (newProps.user !== this.props.user || newProps.device !== this.props.device) {
            setDeviceTitle(newProps.user, newProps.device);
        }
    }

    render() {

        if (this.props.error != null) {
            return (<Error err={this.props.error}/>);
        }
        if (this.props.user == null || this.props.device == null || this.props.streamarray == null) {
            // Currently querying
            return (<Loading/>);
        }

        // React router does not allow using hash routing, so we route by hash here
        switch (this.props.location.hash) {
            case "#create":
                return (<StreamCreate user={this.props.user} device={this.props.device} state={this.props.state.create}/>);
            case "#edit":
                return (<DeviceEdit user={this.props.user} device={this.props.device} state={this.props.state.edit}/>);

        }

        return (<DeviceView user={this.props.user} device={this.props.device} state={this.props.state.view} streamarray={this.props.streamarray}/>);
    }
}

export default connectStorage(connect((store, props) => ({
    state: getDeviceState((props.user != null && props.device != null
        ? props.user.name + "/" + props.device.name
        : ""), store)
}))(Device), false, true);
