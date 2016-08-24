/*
This shows a line chart of the data given
*/

import {addView} from '../datatypes';
import {generateLineChart} from './components/LineChart';
import {generateDropdownLineChart, generateTimeOptions} from './components/DropdownLineChart';
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
    }, {
        ...generateDropdownLineChart("This view averages the datapoint values over the chosen time period", generateTimeOptions("Average", "", "mean"), 1),
        key: "averagedLineView",
        title: "Averaged Values",
        subtitle: ""
    }, {
        ...generateDropdownLineChart("This view sums the datapoint values over the chosen time period", generateTimeOptions("Sum", "", "sum"), 1),
        key: "summedLineView",
        title: "Summed Values",
        subtitle: ""
    }
];

function showLineChart(context) {
    if (context.data.length > 1 && context.pipescript !== null) {

        // We now check if the data is numeric
        if (isNumeric(context.data[0].d) && isNumeric(context.data[context.data.length - 1].d)) {
            return LineView;
        }

    }

    return null;
}

addView(showLineChart);
