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
                <TextField hintText="Nickname" floatingLabelText="Nickname" style={{
                    marginTop: "-20px"
                }} value={this.props.value} onChange={this.props.onChange}/><br/>
            </div>
        );
    }
}
export default NicknameEditor;
