import React, {Component, PropTypes} from 'react';
import TextField from 'material-ui/TextField';

import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import CodeMirror from 'react-codemirror';
import 'codemirror/mode/javascript/javascript';

class SchemaEditor extends Component {
    static propTypes = {
        value: PropTypes.string.isRequired,
        onChange: PropTypes.func.isRequired
    }

    render() {
        return (
            <div style={{
                textAlign: "center"
            }}>
                <h3>JSON Schema</h3>
                <p>The schema that the data within the stream will conform to. This cannot be changed after creating the stream. &nbsp;<a href="http://json-schema.org/examples.html">Learn more here...</a>
                </p>
                <div style={{
                    marginLeft: "auto",
                    marginRight: "auto",
                    border: "1px solid black",
                    width: "80%",
                    textAlign: "left"
                }}><CodeMirror value={this.props.value} onChange={(txt) => this.props.onChange(undefined, txt)} options={{
                mode: "application/json",
                lineWrapping: true
            }}/>
                </div>
            </div>
        );
    }
}
export default SchemaEditor;
