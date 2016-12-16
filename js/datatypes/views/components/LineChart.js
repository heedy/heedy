/*
This shows a line chart of the data given
*/

import React, { PropTypes } from 'react';
import DataTransformUpdater from './DataUpdater';

import { Line } from 'react-chartjs-2';
import moment from 'moment';

class LineChart extends DataTransformUpdater {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        transform: PropTypes.string
    }

    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        let dataset = new Array(d.length);

        // We check if the dataset is boolean - in which case we draw a stepped line
        let isbool = true;

        for (let i = 0; i < d.length; i++) {
            let data = d[i].d;
            if (typeof (data) === "boolean") {
                if (data === false) {
                    data = 0;
                } else {
                    data = 1;
                }
            } else {
                isbool = false;
            }
            dataset[i] = {
                x: moment.unix(d[i].t),
                y: data
            }
        }

        let pointColor = (d.length > 500 ? "rgba(0,0,0,0.1)" : "rgba(0,92,158,0.6)");

        return {
            datasets: [
                {
                    label: name,
                    data: dataset,
                    lineTension: 0,
                    // For nicer displaying, we don't add a fill color when we have enough datapoints,
                    // and when we have a lot of data, we turn into a scatter chart. For booleans, though,
                    // we always use a fill, so that they are more visible.
                    fill: (isbool
                        ? true
                        : d.length < 50),
                    showLine: (isbool
                        ? true
                        : d.length < 500),
                    steppedLine: isbool,
                    backgroundColor: "rgba(66,134,244,0.4)",
                    borderColor: "rgba(0,92,158,0.4)",
                    pointBackgroundColor: pointColor,
                    pointBorderColor: pointColor,
                    pointRadius: (d.length > 500 ? 2 : 3)
                }
            ]
        };
    }

    render() {
        return (<Line data={this.data} options={{
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
            },
            animation: false,
            pointColor: "blue"
        }} />);
    }
}

export default LineChart;

// generate creates a new view that displays a line chart. The view object is set up
// so that it is totally ready to be passed as a result of the shower function
export function generateLineChart(transform) {
    let component = LineChart;

    // If we're given a transform, wrap the LineChart so that we can pass transform into the class.
    if (transform != null) {
        component = React.createClass({
            render: function () {
                return (<LineChart {...this.props} transform={transform} />);
            }
        });
    }

    return { initialState: {}, component: component, width: "expandable-half" };
}
