/*
This shows a bar chart if we deem the data to be amenable to bar-chart-plotting.
The exact conditions are in the function showBarChart below
*/

import {addView} from '../datatypes';
import {generateBarChart} from './components/BarChart';

const BarView = [
    {
        ...generateBarChart("map($,count)", "Counts the occurences of the given values"),
        key: "barView",
        title: "Value Counts",
        subtitle: ""
    }
];

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function isValidKey(n) {
    return (isNumeric(n) || (typeof n) === "string");
}

function showBarChart(context) {
    if (context.data.length < 5 || context.pipescript === null) {
        return null;
    }

    // We unfortunately have to guess whether the data is in a format
    // that can be directly exploited in terms of a bar Chart
    // We use a simple heuristic:
    //    Check the first and last 100 datapoints. Compute the ratio of repetitions/points
    //    If the ratio is greater than a chosen number, and the total percentage of independent (1s)
    //    values is not too high, we display the bar chart.
    let totalpoints = 0;
    let uniquepoints = 0;
    let d = context.data;
    if (d.length < 200) {
        let kv = {};
        for (let i = 0; i < d.length; i++) {
            if (!isValidKey(d[i].d)) {
                return null;
            }
            if (kv[d[i].d] === undefined) {
                uniquepoints += 1;
                kv[d[i].d] = 1;
            }

        }
        totalpoints = d.length;
    } else {
        totalpoints = 200;
        let kv = {}
        for (let i = 0; i < 100; i++) {
            if (!isValidKey(d[i].d)) {
                return null;
            }
            if (kv[d[i].d] === undefined) {
                uniquepoints += 1;
                kv[d[i].d] = 1;
            }
        }
        for (let i = d.length - 100; i < d.length; i++) {
            if (!isValidKey(d[i].d)) {
                return null;
            }
            if (kv[d[i].d] === undefined) {
                uniquepoints += 1;
                kv[d[i].d] = 1;
            }
        }
    }

    if (uniquepoints < totalpoints && uniquepoints < 20 || uniquepoints < 100 && uniquepoints / totalpoints < 0.5) {
        return BarView;
    }

    return null;
}

addView(showBarChart);
