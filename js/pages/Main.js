import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {go} from '../actions';

import MainToolbar from '../components/MainToolbar';

import Welcome from '../components/Welcome';

class DeviceView extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        device: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        streamarray: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        onEditClick: PropTypes.func.isRequired,
        onExpandClick: PropTypes.func.isRequired,
        onAddClick: PropTypes.func.isRequired,
        onStreamClick: PropTypes.func.isRequired
    }

    render() {
        let state = this.props.state;
        let user = this.props.user;
        let device = this.props.device;
        let streams = this.props.streamarray;
        return (
            <div style={{
                textAlign: "left"
            }}>
                <MainToolbar/>
                <Welcome/>
            </div>
        );
    }
}
export default connect(undefined, (dispatch, props) => ({
    onEditClick: () => dispatch(go(props.user.name + "/" + props.device.name + "#edit")),
    onExpandClick: (val) => dispatch({
        type: 'DEVICE_VIEW_EXPANDED',
        name: props.user.name + "/" + props.device.name,
        value: val
    }),
    onAddClick: () => dispatch(go(props.user.name + "/" + props.device.name + "#create")),
    onStreamClick: (s) => dispatch(go(s))
}))(DeviceView);
