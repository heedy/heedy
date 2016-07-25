/*
This shows a line chart of the data given
*/

import {addView} from '../datatypes';

import React, {Component, PropTypes} from 'react';

import {Line} from 'react-chartjs';

import moment from 'moment';

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

// chartjs uses a different form for datapoints, so we convert it to the desired
// form :D
function generateDatasetFromData(d) {
    let dataset = new Array(d.length);

    for (let i = 0; i < d.length; i++) {
        dataset[i] = {
            x: moment.unix(d[i].t),
            y: d[i].d
        }
    }

    return dataset;
}

function getLineDataset(d, name) {
    return {
        datasets: [
            {
                label: name,
                data: generateDatasetFromData(d),
                lineTension: 0
            }
        ]
    };
}

class LineChart extends Component {

    componentWillMount() {
        this.setState({
            data: getLineDataset(this.props.data, this.props.stream.name)
        });
    }

    componentWillReceiveProps(p) {
        if (p.data !== this.props.data) {
            this.setState({
                data: getLineDataset(p.data, p.stream.name)
            });
        }
    }

    render() {

        return (<Line data={this.state.data} options={{
            legend: {
                display: false
            },
            scales: {
                xAxes: [
                    {
                        type: 'time',
                        position: 'bottom'
                    }
                ]
            }
        }}/>);
    }
}

const LineView = {
    key: "lineView",
    component: LineChart,
    width: "expandable-half",
    initialState: {},
    title: "Line Plot",
    subtitle: ""
}

function showLineChart(context) {
    if (context.data.length > 0) {
        // Check the schema
        if (context.schema.type !== undefined && context.schema.type == "numeric") {
            return LineView;
        }
        // We now check if the data is numeric
        if (isNumeric(context.data[0].d) && isNumeric(context.data[context.data.length - 1].d)) {
            return LineView;
        }

    }

    return null;
}

addView(showLineChart);
