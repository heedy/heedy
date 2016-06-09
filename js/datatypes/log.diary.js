import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import datatypes from './datatypes'

import Textarea from 'react-textarea-autosize';
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
                <Textarea style={{
                    width: "100%",
                    fontSize: 18,
                    borderColor: "#ccc"
                }} value={value} minRows={4} useCacheForDOMMeasurements name={this.props.path} multiLine={true} onChange={(e) => this.props.onChange({value: e.target.value})}/><br/>
                <RaisedButton primary={true} label="Submit" onTouchTap={() => this.props.onSubmit(value)}/>
            </div>
        );
    }
}

// register the datatype
datatypes["log.diary"] = {
    input: {
        component: DataInput,
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
