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

import StreamCard from '../components/StreamCard';
import DataTable from '../components/DataTable';
import DataInput from '../components/DataInput';

class StreamView extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        device: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        stream: PropTypes.object.isRequired,
        state: PropTypes.shape({expanded: PropTypes.bool.isRequired}).isRequired,
        thisUser: PropTypes.object.isRequired,
        thisDevice: PropTypes.object.isRequired
    }
    render() {
        let state = this.props.state;
        let user = this.props.user;
        let device = this.props.device;
        let stream = this.props.stream;
        return (
            <div>
                <StreamCard user={user} device={device} stream={stream} state={state}/>

                <div style={{
                    marginLeft: "-15px",
                    marginRight: "-15px"
                }}>
                    {stream.downlink || this.props.thisUser.name == user.name && this.props.thisDevice.name == device.name
                        ? (<DataInput user={user} device={device} stream={stream}/>)
                        : null}

                    <DataTable data={[
                        {
                            timestamp: 34534,
                            data: 45
                        }, {
                            timestamp: 435345345,
                            data: 67
                        }
                    ]}/>
                </div>
            </div>
        );
    }
}

export default connect((state) => ({thisUser: state.site.thisUser, thisDevice: state.site.thisDevice}), (dispatch, props) => ({
    onInputSubmit: (val) => dispatch({
        type: 'STREAM_INPUT_SUBMIT',
        name: props.user.name + "/" + props.device.name + "/" + props.stream.name,
        value: val
    })
}))(StreamView);
