/*
This shows a bar chart, with the option of showing a pie chart
*/

import React, { PropTypes } from 'react';
import DataTransformUpdater from './DataUpdater';
import dropdownTransformDisplay from './dropdownTransformDisplay';
import {dataKeyCompare} from '../typecheck';

import { Bar, Pie } from 'react-chartjs-2';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import moment from 'moment';


// http://stackoverflow.com/questions/25594478/different-color-for-each-bar-in-a-bar-chart-chartjs
function getRandomColor() {
    var letters = '0123456789ABCDEF'.split('');
    var color = '#';
    for (var i = 0; i < 6; i++) {
        color += letters[Math.floor(Math.random() * 16)];
    }
    return color;
}


class BarChart extends DataTransformUpdater {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        transform: PropTypes.string,
        state: PropTypes.object.isRequired
    }

    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        // We assume that we are given a single datapoint, which is the map of
        // key: numeric value for all keys
        if (d.length != 1) {
            console.error("Bar Chart requires a single datapoint");
            return {};
        }

        let keys = Object.keys(d[0].d).sort(dataKeyCompare.bind(d[0].d));
        let data = keys.map((k) => d[0].d[k]);
        let colors = keys.map((k) => getRandomColor());

        return {
            labels: keys,
            datasets: [
                {
                    data: data,
                    backgroundColor: colors,
                    hoverBackgroundColor: colors
                }
            ]
        };

    }

    render() {
        let ispiechart = (this.props.state.piechart !== undefined && this.props.state.piechart === true);

        if (ispiechart) {
            return (<Pie data={this.data} options={{}} />);
        }

        return (<Bar data={this.data} options={{
            scales: {
                xAxes: [{
                    scaleLabel: {
                        display: this.props.xlabel != "",
                        labelString: this.props.xlabel !== undefined ? this.props.xlabel : ""
                    }
                }],
                yAxes: [{
                    ticks: {
                        beginAtZero: true
                    },
                    scaleLabel: {
                        display: this.props.ylabel != "",
                        labelString: this.props.ylabel !== undefined ? this.props.ylabel : ""
                    }
                }]
            },
            legend: {
                display: false
            }
        }} />);
    }
}

export default BarChart;

export function getBarChartIcons(context) {
    if (context.state.piechart !== undefined && context.state.piechart === true) {
        return [(
            <IconButton key="charttype" onTouchTap={() => context.setState({ piechart: false })} tooltip="Bar Chart">
                <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                    insert_chart
                    </FontIcon>
            </IconButton>
        )];
    }
    return [(
        <IconButton key="charttype" onTouchTap={() => context.setState({ piechart: true })} tooltip="Pie Chart">
            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                pie_chart
                </FontIcon>
        </IconButton>
    )];
}

// generate creates a new view that displays a bar chart. The view object is set up
// so that it is totally ready to be passed as a result of the shower function
export function generateBarChart(transform, description, xlabel = "", ylabel = "") {
    let component = BarChart;
    // If we're given a transform, wrap the BarChart so that we can pass transform into the class.
    if (transform != null && transform !== undefined) {
        component = React.createClass({
            render: function () {
                return (<BarChart {...this.props} transform={transform} xlabel={xlabel} ylabel={ylabel} />);
            }
        });
    }

    let result = {
        initialState: {},
        component: component,
        width: "expandable-half",
        icons: getBarChartIcons
    };
    if (transform != null && transform !== undefined) {
        result.dropdown = dropdownTransformDisplay(description, transform);
    }
    return result;
}
