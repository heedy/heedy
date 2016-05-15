import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import Subheader from 'material-ui/Subheader';

import {getUserState} from './state';
import connectStorage from './connectStorage';

import Error from './components/Error';
import Loading from './components/Loading';
import UserCard from './components/UserCard'

class User extends Component {
    static propTypes = {
        user: PropTypes.object,
        error: PropTypes.object,
        editing: PropTypes.bool.isRequired,
        expanded: PropTypes.bool.isRequired,
        onEditClick: PropTypes.func.isRequired,
        onExpand: PropTypes.func.isRequired
    };

    render() {
        if (this.props.error != null) {
            // There was an error
            return (<Error err={this.state.error}/>);
        }
        if (this.props.user == null) {
            // The user is currently being queried
            return (<Loading/>);
        }
        return (
            <div>
                <UserCard user={this.props.user} editing={this.props.editing} onEditClick={this.props.onEditClick} expanded={this.props.expanded} onExpandClick={this.props.onExpand}/>
                <Subheader style={{
                    marginTop: 20
                }}>Devices</Subheader>

            </div>
        );
    }
}
export default connectStorage(connect((store, props) => getUserState((props.user != null
    ? props.user.name
    : ""), store), (dispatch, props) => {
    return {
        onEditClick: (val) => (dispatch({type: "USERPAGE_EDIT", name: props.user.name, value: val})),
        onExpand: (val) => (dispatch({type: "USERPAGE_EXPAND", name: props.user.name, value: val}))
    };
})(User));
