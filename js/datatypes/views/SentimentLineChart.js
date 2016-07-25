/*
This shows a line chart of the sentiment in text!

This is a copy of LineChart.js with just changes to getting sentiment
*/

import {addView} from '../datatypes';

import React, {Component, PropTypes} from 'react';

import {Line} from 'react-chartjs';

import moment from 'moment';

import sentiment from 'sentiment';

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
            y: sentiment(d[i].d).comparative
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
                lineTension: 0,
                fill: false
            }
        ]
    };
}

class LineChart extends Component {

    componentWillMount() {
        this.setState({
            data: getLineDataset(this.props.data, "sentiment")
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
    key: "sentimentView",
    component: LineChart,
    width: "expandable-full",
    initialState: {},
    title: "Text Sentiment",
    subtitle: ""
}

function showLineChart(context) {
    if (context.data.length > 0) {
        if (context.stream.datatype == "log.diary") {
            return LineView;
        }
    }

    return null;
}

addView(showLineChart);
