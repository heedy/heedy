import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {go} from '../actions';

import ObjectCard from '../components/ObjectCard';

class StreamView extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        device: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        stream: PropTypes.object.isRequired,
        state: PropTypes.shape({expanded: PropTypes.bool.isRequired}).isRequired,
        onEditClick: PropTypes.func.isRequired,
        onExpandClick: PropTypes.func.isRequired
    }
    render() {
        let state = this.props.state;
        let user = this.props.user;
        let device = this.props.device;
        let stream = this.props.stream;
        return (
            <div>
                <ObjectCard expanded={state.expanded} onEditClick={this.props.onEditClick} onExpandClick={this.props.onExpandClick} style={{
                    textAlign: "left"
                }} object={stream} path={user.name + "/" + device.name + "/" + stream.name}></ObjectCard>
            </div>
        );
    }
}

export default connect(undefined, (dispatch, props) => ({
    onEditClick: () => dispatch(go(props.user.name + "/" + props.device.name + "/" + props.stream.name + "#edit")),
    onExpandClick: (val) => dispatch({
        type: 'STREAM_VIEW_EXPANDED',
        name: props.user.name + "/" + props.device.name + "/" + props.stream.name,
        value: val
    }),
    onAddClick: () => dispatch(go(props.user.name + "/" + props.device.name + "/" + props.stream.name + "#create")),
    onStreamClick: (s) => dispatch(go(s))
}))(StreamView);
