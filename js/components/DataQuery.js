import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import FlatButton from 'material-ui/FlatButton';
import TextField from 'material-ui/TextField';

import 'bootstrap-daterangepicker/daterangepicker.css';
import DateRangePicker from 'react-bootstrap-daterangepicker';
import moment from 'moment';
import Textarea from 'react-textarea-autosize';

import ExpandableCard from './ExpandableCard';

import {query, showMessage} from '../actions';

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

    static defaultProps = {
        timeranges: {
            'Today': [
                moment().startOf('day'), moment()
            ],
            'Yesterday': [
                moment().subtract(1, 'days').startOf('day'),
                moment().subtract(1, 'days').endOf('day')
            ],
            'Last 7 Days': [
                moment().subtract(7, 'days'),
                moment()
            ],
            'Last 30 Days': [
                moment().subtract(30, 'days'),
                moment()
            ],
            'This Month': [
                moment().startOf('month'), moment().endOf('month')
            ],
            'Last Month': [
                moment().subtract(1, 'month').startOf('month'),
                moment().subtract(1, 'month').endOf('month')
            ]
        }

    }

    componentDidMount() {
        // Set up the data query if we don't have any data
        if (this.props.state.data.length <= 0) {
            this.getDefault();
        }
    }

    getDefault() {
        this.props.query({i1: -50, i2: 0});
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
        this.props.query({bytime: true, t1: s.t1.unix(), t2: s.t2.unix(), limit: s.limit, transform: s.transform});
    }

    render() {
        let state = this.props.state;
        let setState = this.props.setState;

        var start = state.t1.format('YYYY-MM-DD hh:mm:ss a');
        var end = state.t2.format('YYYY-MM-DD hh:mm:ss a');
        var label = start + ' ➡ ' + end;

        return (
            <ExpandableCard state={state} width="expandable-half" setState={this.props.setState} title="Query Data" subtitle="Choose what data is displayed">
                <h5>Time Range</h5>
                <DateRangePicker startDate={state.t1} endDate={state.t2} ranges={this.props.timeranges} opens="left" timePicker={true} onEvent={(e, picker) => setState({t1: picker.startDate, t2: picker.endDate})}>
                    <div id="reportrange" className="selected-date-range-btn" style={{
                        background: "#fff",
                        cursor: "pointer",
                        padding: "5px 10px",
                        border: "1px solid #ccc",
                        width: "100%"
                    }}>
                        <i className="glyphicon glyphicon-calendar fa fa-calendar"></i>&nbsp;

                        <span className="pull-right">{label}</span>
                    </div>
                </DateRangePicker>
                <h5>Transform</h5>
                <Textarea style={{
                    width: "100%",
                    borderColor: "#ccc",
                    fontFamily: "Courier New",
                    fontSize: "17px",
                    padding: "3px"
                }} value={state.transform} minRows={1} useCacheForDOMMeasurements name={this.props.path} multiLine={true} onChange={(event) => setState({transform: event.target.value})}/>

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
            </ExpandableCard>
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
})(DataQuery);
