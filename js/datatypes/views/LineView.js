/*
This shows a line chart of the data given
*/

import {addView} from '../datatypes';
import {generateLineChart} from './components/LineChart';
import dropdownTransformDisplay from './components/dropdownTransformDisplay';

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

const LineView = [
    {
        ...generateLineChart(),
        key: "lineView",
        title: "Line Plot",
        subtitle: ""
    }
];

function showLineChart(context) {
    if (context.data.length > 1) {

        // We now check if the data is numeric
        if (isNumeric(context.data[0].d) && isNumeric(context.data[context.data.length - 1].d)) {
            return LineView;
        }

    }

    return null;
}

addView(showLineChart);
