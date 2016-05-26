import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import FlatButton from 'material-ui/FlatButton';

import TimePicker from 'material-ui/TimePicker';
import TextField from 'material-ui/TextField';

import DataTable from './DataTable';

import {query} from '../actions';

class DataView extends Component {
    static propTypes = {
        state: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        query: PropTypes.func.isRequired
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
        // We now run the query
        this.props.query({bytime: true, t1: 0, t2: 0, limit: 50, transform: this.props.state.transform});
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
                        <TimePicker format="ampm" hintText="Start Time"/>
                        <TimePicker format="ampm" hintText="End Time"/>
                        <TextField fullWidth={true} hintText="PipeScript" floatingLabelText="Transform" style={{
                            marginTop: "-20px"
                        }} value={state.transform} onChange={(val, txt) => setState({
                            ...state,
                            transform: txt
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
        setState: (s) => dispatch({type: "STREAM_VIEW_SET", name: path, value: s})
    };
})(DataView);
