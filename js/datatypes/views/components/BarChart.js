/*
This shows a bar chart, with the option of showing a pie chart
*/

import React, { PropTypes } from 'react';
import DataTransformUpdater from './DataUpdater';
import dropdownTransformDisplay from './dropdownTransformDisplay';

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


// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

// This is a custom comparison function used to sort the keys in increasing order.
// We order things as follows:
//  - If we think that both keys are in a similar format, and have floats in them, sort by the float.
//  - Otherwise, perform a normal compare
var floatmatcher = /[+-]?\d+(\.\d+)?/g;
function dataKeyCompare(a, b) {
    // We first try to extract a number from both strings
    // http://stackoverflow.com/questions/17374893/how-to-extract-floating-numbers-from-strings-in-javascript
    let numa = a.match(floatmatcher)
    if (numa != null && numa.length > 0) {
        let numb = b.match(floatmatcher)
        if (numb != null && numa.length == numb.length) {
            let na = parseFloat(numa[0]);
            let nb = parseFloat(numb[0]);
            return (na < nb
                ? -1
                : (na == nb
                    ? 0
                    : 1));
        }
    }

    // Since we couldn't extract a number, try to match the data values
    if (isNumeric(this[a]) && isNumeric(this[b])) {
        a = this[a];
        b = this[b];
    }

    // Otherwise, return just normal string compare
    return (a > b
        ? -1
        : (a == b
            ? 0
            : 1));
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
                yAxes: [
                    {
                        ticks: {
                            beginAtZero: true
                        }
                    }
                ]
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
export function generateBarChart(transform, description) {
    let component = BarChart;

    // If we're given a transform, wrap the BarChart so that we can pass transform into the class.
    if (transform != null) {
        component = React.createClass({
            render: function () {
                return (<BarChart {...this.props} transform={transform} />);
            }
        });
    }

    let result = {
        initialState: {},
        component: component,
        width: "expandable-half",
        icons: getBarChartIcons
    };
    if (transform != null) {
        result.dropdown = dropdownTransformDisplay(description, transform);
    }
    return result;
}
