import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

class Device extends Component {
    static propTypes = {
        params: PropTypes.shape({user: PropTypes.string.isRequired, device: PropTypes.string.isRequired}).isRequired
    };

    render() {
        return (

            <h1>Device: {this.props.params.user}/{this.props.params.device}</h1>

        );
    }
}

export default Device;
