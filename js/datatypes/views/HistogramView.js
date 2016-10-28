/*
This view shows a histogram of numeric values, binned by a custom bin size
*/

import {addView} from '../datatypes';
import {generateDropdownBarChart} from './components/DropdownBarChart';

const HistView = [
    {
        ...generateDropdownBarChart("This view creates a histogram of your numeric values, with each bar being the bucket width.", [
            {
                name: "Bar Size: 5",
                transform: "map(bucket(5),count)"
            }, {
                name: "Bar Size: 10",
                transform: "map(bucket(10),count)"
            }, {
                name: "Bar Size: 20",
                transform: "map(bucket(20),count)"
            }, {
                name: "Bar Size: 50",
                transform: "map(bucket(50),count)"
            }, {
                name: "Bar Size: 100",
                transform: "map(bucket(100),count)"
            }, {
                name: "Bar Size: 500",
                transform: "map(bucket(500),count)"
            }, {
                name: "Bar Size: 1000",
                transform: "map(bucket(1000),count)"
            }
        ], 1),
        key: "histView",
        title: "Histogram",
        subtitle: ""
    }
];

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function showHistogramView(context) {
    let d = context.data;
    if (d.length < 7 || context.pipescript === null) {
        return null;
    }

    // Try the first 20 and last 20 datapoints to check if they are numeric
    for (let i = 0; i < d.length && i < 20; i++) {
        if (!isNumeric(d[i].d)) {
            return null;
        }
    }
    for (let i = d.length - 1; i >= 20 && i > d.length - 20; i--) {
        if (!isNumeric(d[i].d)) {
            return null;
        }
    }

    return HistView;
}

addView(showHistogramView);
