import React, { Component } from 'react';

import 'bootstrap-daterangepicker/daterangepicker.css';
import DateRangePicker from 'react-bootstrap-daterangepicker';
import moment from 'moment';

const defaultTimeRanges = {
    'Today': [
        moment().startOf('day'), moment().endOf('day')
    ],
    'Yesterday': [
        moment().subtract(1, 'days').startOf('day'),
        moment().subtract(1, 'days').endOf('day')
    ],
    'Last 7 days': [
        moment().subtract(7, 'days'),
        moment().endOf('day')
    ],
    'This Month': [
        moment().startOf('month'), moment().endOf('month')
    ],
    'Last 30 Days': [
        moment().subtract(30, 'days'),
        moment().endOf('day')
    ],
    'Last 3 Months': [
        moment().subtract(3, 'months'),
        moment().endOf('day')
    ],
    'Last Year': [
        moment().subtract(1, 'year'),
        moment().endOf('day')
    ]
};
const timeformat = 'YYYY-MM-DD hh:mm:ss a';

export const TimePicker = ({ state, setState }) => (
    <div>
        <h5>Time Range
            {state.bytime !== undefined ? (<a className="pull-right" style={{
                cursor: "pointer"
            }} onClick={() => setState({ bytime: false })}>Switch to index range</a>) : null}
        </h5>
        <DateRangePicker startDate={state.t1} endDate={state.t2} ranges={defaultTimeRanges} opens="left" timePicker={true} onEvent={(e, picker) => setState({ t1: picker.startDate, t2: picker.endDate })}>
            <div id="reportrange" className="selected-date-range-btn" style={{
                background: "#fff",
                cursor: "pointer",
                padding: "5px 10px",
                border: "1px solid #ccc",
                width: "100%",
                textAlign: "center"
            }}>
                <i className="glyphicon glyphicon-calendar fa fa-calendar pull-right"></i>&nbsp;
                <span>{state.t2.format(timeformat)}&nbsp;&nbsp;&nbsp;&nbsp;{' ➡ '}&nbsp;&nbsp;&nbsp;&nbsp;{state.t1.format(timeformat)}</span>
            </div>
        </DateRangePicker>
    </div>
);

export const IndexPicker = ({ state, setState }) => (
    <div>
        <h5>Index Range
            {state.bytime !== undefined ? (<a className="pull-right" style={{
                cursor: "pointer"
            }} onClick={() => setState({ bytime: true })}>Switch to time range</a>) : null}
        </h5>
        <h6>Remember that negative values represent values from the data's end: [-50,0) will give the most recent 50 datapoints.</h6>
        <div style={{
            textAlign: "center"
        }}>
            <input value={state.i1} type="number" className="pull-left" style={{
                textAlign: "center"
            }} onChange={(e) => setState({ i1: e.target.value })} />
            <input type="number" className="pull-right" style={{
                textAlign: "center"
            }} value={state.i2} onChange={(e) => setState({ i2: e.target.value })} />
            <span>{' ➡ '}</span>
        </div>
    </div>
);

const QueryRange = ({ state, setState }) => (state.bytime ?
    (<TimePicker state={state} setState={setState} />)
    : (<IndexPicker state={state} setState={setState} />)
);

export default QueryRange;