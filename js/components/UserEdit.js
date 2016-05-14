import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

class UserEdit extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired
    }

    render() {

        return (
            <div>
                Editing
            </div>
        );
    }
}

export default UserEdit;
