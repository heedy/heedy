import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {RadioButton, RadioButtonGroup} from 'material-ui/RadioButton';

import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import CodeMirror from 'react-codemirror';
import 'codemirror/mode/javascript/javascript';

class SchemaEditor extends Component {
    static propTypes = {
        value: PropTypes.string.isRequired,
        defaultSchemas: PropTypes.arrayOf(PropTypes.object).isRequired,
        onChange: PropTypes.func.isRequired
    }

    render() {

        return (
            <div>
                <h3>JSON Schema</h3>
                <p>The schema that the data within the stream will conform to. This cannot be changed after creating the stream. &nbsp;<a href="http://json-schema.org/examples.html">Learn more here...</a>
                </p>
                <RadioButtonGroup name="schema" valueSelected={this.props.value} onChange={this.props.onChange}>
                    {this.props.defaultSchemas.map((s) => {
                        let schemaString = JSON.stringify(s.schema);
                        return (<RadioButton value={schemaString} key={schemaString} label={s.description}/>);
                    })}
                </RadioButtonGroup>
                <div style={{
                    marginTop: "10px",
                    border: "1px solid black",
                    width: "80%",
                    maxWidth: "300px",
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
export default connect((state) => ({defaultSchemas: state.site.defaultschemas}))(SchemaEditor);
