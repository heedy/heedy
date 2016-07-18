import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import ExpandableCard from './ExpandableCard';

import {dataInput, showMessage} from '../actions';
import {getInput} from '../datatypes/datatypes';
import {getStreamState} from '../reducers/stream';

class DataInput extends Component {
    static propTypes = {
        state: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        onSubmit: PropTypes.func.isRequired,
        setState: PropTypes.func.isRequired,
        showMessage: PropTypes.func.isRequired,
        title: PropTypes.string,
        subtitle: PropTypes.string
    }

    static defaultProps = {
        title: "Insert Into Stream",
        subtitle: ""
    }

    touch() {
        if (this.props.touch !== undefined) {
            this.props.touch();
        }
    }
    render() {
        let user = this.props.user;
        let device = this.props.device;
        let stream = this.props.stream;
        let path = user.name + "/" + device.name + "/" + stream.name;

        let state = this.props.state;

        let schema = JSON.parse(stream.schema);

        // Based on the stream datatype, we get the component to display as an input. The default being
        // a standard form, but based on the datatype, it can be stars, or whatever is desired.
        let datatype = getInput(stream.datatype);

        return (
            <ExpandableCard state={state} width={datatype.width} setState={this.props.setState} width={datatype.width} title={this.props.title} subtitle={this.props.subtitle} dropdown={datatype.dropdown}>
                <datatype.component {...this.props} schema={schema} path={path}/>
            </ExpandableCard>
        );
    }
}

export default connect((state, props) => ({
    state: getStreamState(props.user.name + "/" + props.device.name + "/" + props.stream.name, state).input
}), (dispatch, props) => ({
    onSubmit: (val, cng) => dispatch(dataInput(props.user, props.device, props.stream, val, cng)),
    showMessage: (val) => dispatch(showMessage(val)),
    setState: (v) => dispatch({
        type: "STREAM_INPUT",
        name: props.user.name + "/" + props.device.name + "/" + props.stream.name,
        value: v
    })

}))(DataInput);
