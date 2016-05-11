import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

class User extends Component {
    static propTypes = {
        params: PropTypes.shape({user: PropTypes.string.isRequired}).isRequired
    };

    render() {
        return (

            <h1>User: {this.props.params.user}</h1>

        );
    }
}

export default User;
