import React, {Component, PropTypes} from 'react';
import Checkbox from 'material-ui/Checkbox';

class VisibleEditor extends Component {
    static propTypes = {
        value: PropTypes.bool.isRequired,
        onChange: PropTypes.func.isRequired,
        type: PropTypes.string.isRequired
    }

    render() {
        return (
            <div>
                <h3>Visible</h3>
                <p>Whether or not the {this.props.type + " "}
                    is shown when listing devices in the web interface. This is to reduce clutter - the device will still be available over REST.</p>
                <Checkbox label="Visible" checked={this.props.value} onCheck={this.props.onChange}/>
            </div>
        );
    }
}
export default VisibleEditor;
