/*
This shows a line chart of the data given
*/

import { addView } from '../datatypes';
import { generateLineChart, LineChart } from './components/LineChart';
import { generateDropdownLineChart, generateTimeOptions } from './components/DropdownLineChart';
import dropdownTransformDisplay from './components/dropdownTransformDisplay';

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function lineViewGenerator(pretransform) {
    return [
        {
            ...generateLineChart(),
            key: "lineView",
            title: "Line Plot",
            subtitle: ""
        }, {
            ...generateDropdownLineChart("This view averages the datapoint values over the chosen time period", generateTimeOptions("Average", "", pretransform + "mean"), 1),
            key: "averagedLineView",
            title: "Averaged Values",
            subtitle: ""
        }, {
            ...generateDropdownLineChart("This view sums the datapoint values over the chosen time period", generateTimeOptions("Sum", "", pretransform + "sum"), 1),
            key: "summedLineView",
            title: "Summed Values",
            subtitle: ""
        }
    ];
}

const LineView = lineViewGenerator("");

// We can visualize objects if their components are numeric
const ObjectLineView = [LineView[0]];

function showLineChart(context) {
    if (context.data.length > 1 && context.pipescript !== null) {

        // We now check if the data is numeric
        if (isNumeric(context.data[0].d) && isNumeric(context.data[context.data.length - 1].d)) {
            return LineView;
        } else if (context.data[0].d !== null && typeof context.data[0].d === 'object'
            && Object.keys(context.data[0].d).length <= LineChart.objectColors.length
            && typeof context.data[context.data.length - 1].d === 'object') {
            // There are object datapoints. Let's make sure the keys match and the values are all numeric.
            // If they do, we can display it as a multiple line series
            let d0keys = Object.keys(context.data[0].d);
            if (d0keys.length == Object.keys(context.data[context.data.length - 1].d).length) {
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
                // Return a object line view if the keys are OK, AND it it isn't a GPS datapoint (latitude/longitude)
                if (keysok && context.data[0].d["latitude"] === undefined) {
                    if (d0keys.length == 1) {
                        // If there is only one key, then we can show this easily
                        return lineViewGenerator(`$('${d0keys[0]}'):`);
                    }
                    return ObjectLineView;
                }
            }
        }

    }

    return null;
}

addView(showLineChart);
