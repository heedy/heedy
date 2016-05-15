import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import Subheader from 'material-ui/Subheader';

import connectStorage from './connectStorage';

import Error from './components/Error';
import Loading from './components/Loading';
import UserCard from './components/UserCard'

class User extends Component {
    static propTypes = {
        user: PropTypes.object,
        error: PropTypes.object
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
                <UserCard user={this.props.user} editing={false} onEditClick={() => {
                    console.log("edit click");
                }}/>
                <Subheader style={{
                    marginTop: 20
                }}>Devices</Subheader>

            </div>
        );
    }
}
export default connectStorage(User);
