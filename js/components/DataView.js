import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import FlatButton from 'material-ui/FlatButton';

import TextField from 'material-ui/TextField';

import DataTable from './DataTable';

import DateTime from 'react-datetime';
import 'react-datetime/css/react-datetime.css';

import {query, showMessage} from '../actions';

class DataView extends Component {
    static propTypes = {
        state: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        query: PropTypes.func.isRequired,
        msg: PropTypes.func.isRequired
    }

    componentDidMount() {
        // Set up the data query if we don't have any data
        if (this.props.state.data.length <= 0) {
            this.getDefault();
        }
    }

    getDefault() {
        this.props.query({i1: -5, i2: 0});
    }

    query() {
        let s = this.props.state;
        if (typeof s.t1 === 'string' || s.t1 instanceof String) {
            this.props.msg("Start Time Invalid");
            return;
        }
        if (typeof s.t2 === 'string' || s.t2 instanceof String) {
            this.props.msg("End Time Invalid");
            return;
        }
        // We now run the query
        this.props.query({bytime: true, t1: s.t1.unix(), t2: s.t2.unix(), limit: 50, transform: s.transform});
    }

    render() {
        let state = this.props.state;
        let setState = this.props.setState;
        return (
            <div className={state.fullwidth
                ? "col-lg-12"
                : "col-lg-6"}>
                <Card style={{
                    marginTop: "20px",
                    textAlign: "left"
                }} onExpandChange={(val) => setState({
                    ...state,
                    tExpanded: val
                })} expanded={state.tExpanded}>
                    <CardHeader title={"Most Recent Data"} showExpandableButton={true}>
                        <div style={{
                            float: "right",
                            marginRight: 25,
                            marginTop: "-15px",
                            marginLeft: "-100px"
                        }}>
                            <IconButton onTouchTap= { (val) => this.getDefault() } tooltip="Get most recent 5 datapoints">
                                <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                    refresh
                                </FontIcon>
                            </IconButton>
                            {state.fullwidth
                                ? (
                                    <IconButton onTouchTap= { (val) => setState({ ...state, fullwidth: false }) }>
                                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                            call_received
                                        </FontIcon>
                                    </IconButton>

                                )
                                : (

                                    <IconButton onTouchTap= { (val) => setState({ ...state, fullwidth: true }) }>
                                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                            call_made
                                        </FontIcon >
                                    </IconButton>
                                )}
                        </div>
                    </CardHeader>
                    <CardText expandable={true} style={{
                        backgroundColor: "rgba(0,179,74,0.05)",
                        paddingBottom: "30px"
                    }}>
                        <p>Query the stream's data starting from the start time and ending at the end time. A maximum of 50 datapoints will be shown.</p>
                        <h5>Start Time</h5>
                        <DateTime onChange={(d) => {
                            setState({
                                ...state,
                                t1: d
                            });
                        }}/>
                        <h5>End Time</h5>
                        <DateTime onChange={(d) => {
                            setState({
                                ...state,
                                t2: d
                            });
                        }}/>
                        <h5>Transform</h5>
                        <input type="text" className="form-control" value={state.transform} onChange={(event) => setState({
                            ...state,
                            transform: event.target.value
                        })}/>
                        <FlatButton style={{
                            float: "right"
                        }} primary={true} label="Run Query" onTouchTap={() => this.query()}/> {state.error !== null
                            ? (
                                <p style={{
                                    paddingTop: "10px"
                                }}>{state.error.msg}</p>
                            )
                            : (
                                <p style={{
                                    paddingTop: "10px"
                                }}>Learn about transforms
                                    <a href="https://connectordb.github.io/pipescript/">{" "}here.</a>
                                </p>
                            )}
                    </CardText>
                    <CardText>
                        <DataTable data={state.data}/>
                    </CardText>
                </Card>
            </div>
        );
    }
}
export default connect(undefined, (dispatch, props) => {
    let path = props.user.name + "/" + props.device.name + "/" + props.stream.name;
    return {
        query: (q) => dispatch(query(props.user, props.device, props.stream, q)),
        setState: (s) => dispatch({type: "STREAM_VIEW_SET", name: path, value: s}),
        msg: (t) => dispatch(showMessage(t))
    };
})(DataView);
