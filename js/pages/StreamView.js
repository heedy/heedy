import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {
    Table,
    TableBody,
    TableHeader,
    TableHeaderColumn,
    TableRow,
    TableRowColumn
} from 'material-ui/Table';

import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import CodeMirror from 'react-codemirror';
import 'codemirror/mode/javascript/javascript';

import TimeDifference from '../components/TimeDifference';
import {go} from '../actions';

import ObjectCard from '../components/ObjectCard';

class StreamView extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        device: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        stream: PropTypes.object.isRequired,
        state: PropTypes.shape({expanded: PropTypes.bool.isRequired}).isRequired,
        onEditClick: PropTypes.func.isRequired,
        onExpandClick: PropTypes.func.isRequired
    }
    render() {
        let state = this.props.state;
        let user = this.props.user;
        let device = this.props.device;
        let stream = this.props.stream;
        return (
            <div>
                <ObjectCard expanded={state.expanded} onEditClick={this.props.onEditClick} onExpandClick={this.props.onExpandClick} style={{
                    textAlign: "left"
                }} object={stream} path={user.name + "/" + device.name + "/" + stream.name}>
                    <Table selectable={false}>
                        <TableHeader enableSelectAll={false} displaySelectAll={false} adjustForCheckbox={false}>
                            <TableRow>
                                <TableHeaderColumn>Datatype</TableHeaderColumn>
                                <TableHeaderColumn>Downlink</TableHeaderColumn>
                                <TableHeaderColumn>Ephemeral</TableHeaderColumn>
                                <TableHeaderColumn>Queried</TableHeaderColumn>
                            </TableRow>
                        </TableHeader>
                        <TableBody displayRowCheckbox={false}>
                            <TableRow>
                                <TableRowColumn>{stream.datatype}</TableRowColumn>
                                <TableRowColumn>{stream.downlink
                                        ? "true"
                                        : "false"}</TableRowColumn>
                                <TableRowColumn>{stream.ephemeral
                                        ? "true"
                                        : "false"}</TableRowColumn>
                                <TableRowColumn><TimeDifference timestamp={stream.timestamp}/></TableRowColumn>
                            </TableRow>
                        </TableBody>
                    </Table >
                    <h4 style={{
                        textAlign: "center"
                    }}>JSON Schema</h4>
                    <div style={{
                        marginLeft: "auto",
                        marginRight: "auto",
                        border: "1px solid black",
                        width: "80%"
                    }}>
                        <CodeMirror value={JSON.stringify(JSON.parse(stream.schema), null, 4)} options={{
                            mode: "application/json",
                            lineWrapping: true,
                            readOnly: true
                        }}/>
                    </div>
                </ObjectCard>

            </div>
        );
    }
}

export default connect(undefined, (dispatch, props) => ({
    onEditClick: () => dispatch(go(props.user.name + "/" + props.device.name + "/" + props.stream.name + "#edit")),
    onExpandClick: (val) => dispatch({
        type: 'STREAM_VIEW_EXPANDED',
        name: props.user.name + "/" + props.device.name + "/" + props.stream.name,
        value: val
    }),
    onAddClick: () => dispatch(go(props.user.name + "/" + props.device.name + "/" + props.stream.name + "#create")),
    onStreamClick: (s) => dispatch(go(s))
}))(StreamView);
