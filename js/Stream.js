import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

class Stream extends Component {
    static propTypes = {
        params: PropTypes.shape({user: PropTypes.string.isRequired, device: PropTypes.string.isRequired, stream: PropTypes.string.isRequired}).isRequired
    };

    render() {
        return (

            <h1>Stream: {this.props.params.user}/{this.props.params.device}/{this.props.params.stream}</h1>

        );
    }
}

export default Stream;
