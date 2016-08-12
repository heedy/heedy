/*
This is the textbox used to input a transform
*/
import React, {Component, PropTypes} from 'react';
import Textarea from 'react-textarea-autosize';

class TransformInput extends Component {
    static propTypes = {
        transform: PropTypes.string.isRequired,
        onChange: PropTypes.func.isRequired
    }
    render() {
        return (<Textarea useCacheForDOMMeasurements value={this.props.transform} minRows={1} style={{
            width: "100%",
            borderColor: "#ccc",
            fontFamily: "Courier New",
            fontSize: "17px",
            padding: "3px"
        }} multiLine={true} onChange={(event) => this.props.onChange(event.target.value)}/>);
    }
}

export default TransformInput;
