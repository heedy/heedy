import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import FlatButton from 'material-ui/FlatButton';
import {
    Table,
    TableBody,
    TableHeader,
    TableHeaderColumn,
    TableRow,
    TableRowColumn
} from 'material-ui/Table';
import TimePicker from 'material-ui/TimePicker';
import TextField from 'material-ui/TextField';

import {query} from '../actions';

class DataTable extends Component {
    static propTypes = {
        state: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        query: PropTypes.func.isRequired
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
                                        </FontIcon>
                                    </IconButton>

                                )}
                        </div>
                    </CardHeader>
                    <CardText expandable={true}>
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
                        }} primary={true} label="Run Query" onTouchTap={() => this.query(state)}/>
                    </CardText>
                    <CardText>
                        <Table selectable={false}>
                            <TableHeader enableSelectAll={false} displaySelectAll={false} adjustForCheckbox={false}>
                                <TableRow>
                                    <TableHeaderColumn>Timestamp</TableHeaderColumn>
                                    <TableHeaderColumn>Data</TableHeaderColumn>
                                </TableRow>
                            </TableHeader>
                            <TableBody displayRowCheckbox={false}>
                                {state.data.map((d) => {
                                    let t = new Date(d.timestamp * 1000);
                                    let ts = t.getHours() + ":" + t.getMinutes() + ":" + t.getSeconds() + " - " + (t.getMonth() + 1) + "/" + t.getDate() + "/" + t.getFullYear();
                                    return (
                                        <TableRow key={JSON.stringify(d)}>
                                            <TableRowColumn>{ts}</TableRowColumn>
                                            <TableRowColumn>{JSON.stringify(d.data)}</TableRowColumn>
                                        </TableRow>
                                    );
                                })}

                            </TableBody>
                        </Table >
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
})(DataTable);
