/*
  This is the main navigator shown for devices. This component chooses the correct page to set based upon the data
  it is getting (shows loading page if the user/device/stream are not ready).

  The component also performs further routing based upon the hash. This is because react-router does not
  support both normal and hash-based routing at the same time.
  All child pages are located in ./pages. This component can be throught of as an extension to the main app routing
  done in App.js, with additional querying for the user/device/stream we want to view.

  It also queries the user/device/stream-specific state from redux, so further children can just use the state without worrying
  about which user/device/stream it belongs to.
*/

import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {getDeviceState} from './reducers/device';
import connectStorage from './connectStorage';

import Error from './components/Error';
import Loading from './components/Loading';

import DeviceView from './pages/DeviceView';
import DeviceEdit from './pages/DeviceEdit';
import StreamCreate from './pages/StreamCreate';
import StreamCreateDatatype from './pages/StreamCreateDatatype';

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
            case "#create-rating":
                return (<StreamCreateDatatype user={this.props.user} device={this.props.device} state={this.props.state.create} datatype="rating.stars"/>)
            case "#create-log":
                return (<StreamCreateDatatype user={this.props.user} device={this.props.device} state={this.props.state.create} datatype="log.diary"/>)

        }

        return (<DeviceView user={this.props.user} device={this.props.device} state={this.props.state.view} streamarray={this.props.streamarray}/>);
    }
}

export default connectStorage(connect((store, props) => ({
    state: getDeviceState((props.user != null && props.device != null
        ? props.user.name + "/" + props.device.name
        : ""), store)
}))(Device), false, true);
