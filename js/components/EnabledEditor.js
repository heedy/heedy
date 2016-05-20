import React, {Component, PropTypes} from 'react';
import Checkbox from 'material-ui/Checkbox';

class EnabledEditor extends Component {
    static propTypes = {
        value: PropTypes.bool.isRequired,
        onChange: PropTypes.func.isRequired,
        type: PropTypes.string.isRequired
    }

    render() {
        return (
            <div>
                <h3>Enabled</h3>
                <p>Whether or not the {this.props.type + " "}
                    accepts data, and is currently functioning.</p>
                <Checkbox label="Enabled" checked={this.props.value} onCheck={this.props.onChange}/>
            </div>
        );
    }
}
export default EnabledEditor;
