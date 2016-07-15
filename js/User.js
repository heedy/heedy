/*
  This is the main navigator shown for users. This component chooses the correct page to set based upon the data
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

import {getUserState} from './reducers/user';
import connectStorage from './connectStorage';

import Error from './components/Error';
import Loading from './components/Loading';

import UserView from './pages/UserView';
import UserEdit from './pages/UserEdit';
import DeviceCreate from './pages/DeviceCreate';

import {setTitle} from './util';

function setUserTitle(user) {
    setTitle(user == null
        ? ""
        : user.name);
}

class User extends Component {
    static propTypes = {
        user: PropTypes.object,
        devarray: PropTypes.object,
        error: PropTypes.object,
        location: PropTypes.object.isRequired,
        state: PropTypes.object
    };
    componentDidMount() {
        setUserTitle(this.props.user);
    }
    componentWillReceiveProps(newProps) {
        if (newProps.user !== this.props.user) {
            setUserTitle(newProps.user);
        }
    }
    render() {
        if (this.props.error != null) {
            return (<Error err={this.props.error}/>);
        }
        if (this.props.user == null || this.props.devarray == null) {
            // Currently querying
            return (<Loading/>);
        }

        // React router does not allow using hash routing, so we route by hash here
        switch (this.props.location.hash) {
            case "#create":
                return (<DeviceCreate user={this.props.user} state={this.props.state.create}/>);
            case "#edit":
                return (<UserEdit user={this.props.user} state={this.props.state.edit}/>);

        }

        return (<UserView user={this.props.user} state={this.props.state.view} devarray={this.props.devarray}/>);
    }
}
export default connectStorage(connect((store, props) => ({
    state: getUserState((props.user != null
        ? props.user.name
        : ""), store)
}))(User), true, false);
