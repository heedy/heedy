import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {getUserState} from './reducers/user';
import connectStorage from './connectStorage';

import Error from './components/Error';
import Loading from './components/Loading';

import UserView from './pages/UserView';
import UserEdit from './pages/UserEdit';
import DeviceCreate from './pages/DeviceCreate';

class User extends Component {
    static propTypes = {
        user: PropTypes.object,
        error: PropTypes.object,
        location: PropTypes.object.isRequired,
        state: PropTypes.object
    };

    render() {
        if (this.props.error != null) {
            return (<Error err={this.props.error}/>);
        }
        if (this.props.user == null) {
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

        return (<UserView user={this.props.user} state={this.props.state.view}/>);
    }
}
export default connectStorage(connect((store, props) => ({
    state: getUserState((props.user != null
        ? props.user.name
        : ""), store)
}))(User));
