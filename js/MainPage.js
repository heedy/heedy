import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {getDeviceState} from './reducers/device';
import connectStorage from './connectStorage';

import Error from './components/Error';
import Loading from './components/Loading';

import Main from './pages/Main';
import StreamCreate from './pages/StreamCreate';

import {setTitle} from './util';

class MainPage extends Component {
    static propTypes = {
        user: PropTypes.object,
        device: PropTypes.object,
        streamarray: PropTypes.object,
        error: PropTypes.object,
        location: PropTypes.object.isRequired,
        state: PropTypes.object
    };
    componentDidMount() {
        setTitle("");
    }
    componentWillReceiveProps(newProps) {
        setTitle("");
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
            case "#addrating":
                return (<StreamCreate user={this.props.user} device={this.props.device} state={this.props.state.create}/>);
        }

        return (<Main user={this.props.user} device={this.props.device} state={this.props.state.view} streamarray={this.props.streamarray}/>);

    }
}

export default connect((store, props) => ({
    state: getDeviceState(store.site.thisUser.name + "/" + store.site.thisDevice.name, store),
    user: store.site.thisUser.name,
    device: store.site.thisDevice.name
}))(connectStorage(MainPage, false, true));
