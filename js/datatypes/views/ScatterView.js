import React, { Component, PropTypes } from 'react';

import DataTransformUpdater from './components/DataUpdater';
import { Line } from 'react-chartjs-2';
import { addView } from '../datatypes';

function fixData(data) {
    if (typeof (data) === "boolean") {
        if (data === false) {
            return 0;
        } else {
            return 1;
        }
    }
    return data;
}

class ScatterComponent extends DataTransformUpdater {

    //The color to give datapoints - when there is less than 500 points, "low" is used
    // and when there are more than 500, "high" is used.
    static pointColor = {
        low: "rgba(75, 192, 192,0.6)",
        high: "rgba(75, 192, 192,0.1)"
    }
    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        if (d.length == 0) {
            return { datasets: [] };
        }

        let keys = Object.keys(d[0].d);
        let dataset = new Array(d.length);
        
        let minColor = 9999999999999;
        let maxColor = -minColor;

        for (let i = 0; i < d.length; i++) {
            dataset[i] = {
                x: fixData(d[i].d[keys[0]]),
                y: fixData(d[i].d[keys[1]]),
            };

            if (keys.length == 3) {
                let dp = fixData(d[i].d[keys[2]]);
                if (dp < minColor) minColor = dp;
                if (dp > maxColor) maxColor = dp;
            }
        }

        let pointColor = (d.length > 500 ? ScatterComponent.pointColor.high : ScatterComponent.pointColor.low);

        if (keys.length == 3 && minColor!=maxColor) {
            pointColor = new Array(d.length);
            for (let i=0; i < d.length; i++) {
                let dp = fixData(d[i].d[keys[2]]);
                pointColor[i] = `hsla(${Math.floor(120 * (dp-minColor)/(maxColor - minColor))},100%,50%,0.4)`;
            }
        }

        return {
            datasets: [{
                label: "",
                data: dataset,
                lineTension: 0,
                fill: false,
                showLine: false,
                pointRadius: (d.length > 500 ? 2 : 3),
                pointBackgroundColor: pointColor,
                pointBorderColor: pointColor,
            }],
            xLabels: [keys[0]],
            yLabels: [keys[0]]
        };
    }
    render() {
        return (<Line data={this.data} options={{
            legend: {
                display: false
            },
            scales: {
                xAxes: [{
                    type: 'linear',
                    position: 'bottom',
                    scaleLabel: {
                        display: true,
                        labelString: Object.keys(this.props.data[0].d)[0]
                    }
                }],
                yAxes: [{
                    type: 'linear',
                    position: 'left',
                    scaleLabel: {
                        display: true,
                        labelString: Object.keys(this.props.data[0].d)[1]
                    }
                }]
            },
            animation: false
        }} />);
    }
}




const ScatterView = {
    key: "scatterView",
    component: ScatterComponent,
    width: "expandable-half",
    title: "Scatter Plot",
    initialState: {},
    subtitle: ""
};

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function showScatterView(context) {
    console.log("Checking ScatterView");
    if (context.data.length > 1
        && context.data[0].d !== null && typeof context.data[0].d === 'object'
        && (Object.keys(context.data[0].d).length == 2 || Object.keys(context.data[0].d).length == 3)
        && typeof context.data[context.data.length - 1].d === 'object') {

        // It looks promising! Let's check that the keys are numeric, to make sure
        // that we can display a scatter chart
        let d0keys = Object.keys(context.data[0].d);
        let keysok = true;
        for (let i = 0; i < d0keys.length; i++) {
            // Make sure the keys match
            if (context.data[context.data.length - 1].d[d0keys[i]] === undefined) {
                keysok = false;
                break;
            }
            // Make sure the contained data is numeric
            if (!isNumeric(context.data[context.data.length - 1].d[d0keys[i]]
                || !isNumeric(context.data[0].d[d0keys[i]]))) {
                keysok = false;
                break;
            }
        }
        // Return the scatter view if the keys are OK, AND it it isn't a GPS datapoint (latitude/longitude)
        if (keysok && context.data[0].d["latitude"] === undefined) {
            return ScatterView;
        }
    }
    return null;
}

addView(showScatterView);