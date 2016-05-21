import React, {Component, PropTypes} from 'react';
import TextField from 'material-ui/TextField';

class DatatypeEditor extends Component {
    static propTypes = {
        value: PropTypes.string.isRequired,
        onChange: PropTypes.func.isRequired,
        schema: PropTypes.string.isRequired
    }

    render() {
        return (
            <div>
                <h3>Datatype</h3>
                <p>A stream's datatype tells ConnectorDB how the data should be interpreted.</p>
                <TextField hintText="number.rating.stars.5" floatingLabelText="Datatype" style={{
                    marginTop: "-20px"
                }} value={this.props.value} onChange={this.props.onChange}/><br/>
            </div>
        );
    }
}
export default DatatypeEditor;
