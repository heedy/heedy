import React, { Component, PropTypes } from 'react';
import { connect } from 'react-redux';
import { Card, CardText, CardHeader } from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import FlatButton from 'material-ui/FlatButton';
import TextField from 'material-ui/TextField';

import QueryRange from './QueryRange';
import TransformInput from './TransformInput';

import ExpandableCard from './ExpandableCard';

import { query, showMessage } from '../actions';

class DataQuery extends Component {
    static propTypes = {
        state: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        query: PropTypes.func.isRequired,
        msg: PropTypes.func.isRequired,
        timeranges: PropTypes.object
    }

    componentDidMount() {
        // Set up the data query if we don't have any data
        if (this.props.state.data.length <= 0) {
            this.query();
        }
    }

    query() {
        let s = this.props.state;
        // We now run the query
        this.props.query({
            bytime: s.bytime,
            i1: parseInt(s.i1),
            i2: parseInt(s.i2),
            t1: s.t1.unix(),
            t2: s.t2.unix(),
            limit: s.limit,
            transform: s.transform
        });
    }

    render() {
        let state = this.props.state;
        let setState = this.props.setState;

        var start = state.t1.format('YYYY-MM-DD hh:mm:ss a');
        var end = state.t2.format('YYYY-MM-DD hh:mm:ss a');
        var label = start + ' ➡ ' + end;

        return (
            <ExpandableCard state={state} width="expandable-half" setState={this.props.setState} title="Query Data" subtitle="Choose what data is displayed">
                <QueryRange state={state} setState={setState} />
                <h5 style={{
                    paddingTop: "10px"
                }}>Server-Side Transform</h5>
                <TransformInput transform={state.transform} onChange={(txt) => setState({ transform: txt })} />

                <FlatButton style={{
                    float: "right"
                }} primary={true} label="Run Query" onTouchTap={() => this.query()} /> {state.error !== null
                    ? (
                        <p style={{
                            paddingTop: "10px",
                            color: "red"
                        }}>{state.error.msg}</p>
                    )
                    : (
                        <p style={{
                            paddingTop: "10px"
                        }}>Learn about transforms
                            <a href="https://connectordb.io/docs/pipescript/">{" "}here.</a>
                        </p>
                    )}
            </ExpandableCard>
        );
    }
}
export default connect(undefined, (dispatch, props) => {
    let path = props.user.name + "/" + props.device.name + "/" + props.stream.name;
    return {
        query: (q) => dispatch(query(props.user, props.device, props.stream, q)),
        setState: (s) => dispatch({ type: "STREAM_VIEW_SET", name: path, value: s }),
        msg: (t) => dispatch(showMessage(t))
    };
})(DataQuery);
