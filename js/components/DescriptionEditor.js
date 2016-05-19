import React, {Component, PropTypes} from 'react';
import TextField from 'material-ui/TextField';

class NicknameEditor extends Component {
    static propTypes = {
        value: PropTypes.string.isRequired,
        onChange: PropTypes.func.isRequired,
        type: PropTypes.string.isRequired
    }

    render() {
        return (
            <div>
                <h3>Description</h3>
                <p>A description can be thought of as a README for the {this.props.type}</p>
                <TextField floatingLabelText="Description" multiLine={true} fullWidth={true} value={this.props.value} style={{
                    marginTop: "-20px"
                }} onChange={this.props.onChange}/><br/>
            </div>
        );
    }
}
export default NicknameEditor;
