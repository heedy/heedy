import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import datatypes from './datatypes'

import {getStreamState} from '../reducers/stream';

import TextField from 'material-ui/TextField';
import RaisedButton from 'material-ui/RaisedButton';

export const diarySchema = {
    type: "string",
    minLength: 1
}

class DataInput extends Component {
    static propTypes = {
        state: PropTypes.object,
        path: PropTypes.string.isRequired,
        onChange: PropTypes.func,
        onSubmit: PropTypes.func
    }
    render() {
        let value = this.props.state.value;
        if (value === undefined || value == null)
            value = "";
        return (
            <div>
                <TextField name={this.props.path} multiLine={true} fullWidth={true} value={value} style={{
                    marginTop: "-20px"
                }} onChange={this.props.onChange}/><br/>
                <RaisedButton primary={true} label="Submit" onTouchTap={() => this.props.onSubmit(value)}/>
            </div>
        );
    }
}

let DIConnected = connect((state, props) => ({
    state: getStreamState(props.path, state).input
}), (dispatch, props) => ({
    onChange: (v, txt) => dispatch({
        type: "STREAM_INPUT",
        name: props.path,
        value: {
            value: txt
        }
    })
}))(DataInput)

// register the datatype
datatypes["log.diary"] = {
    input: {
        component: DIConnected,
        size: 2, // One of 1 or 2 meaning normal or double size of the data input card
    },
    create: {
        required: null,
        optional: null,
        description: "A log (or diary) can be used to write about events in your life. Analysis of the text might reveal general trends in your thoughts or what events are associated with certain ratings.",
        default: {
            schema: JSON.stringify(diarySchema),
            datatype: "log.diary",
            ephemeral: false,
            downlink: false
        }
    },
    name: "log"
};
