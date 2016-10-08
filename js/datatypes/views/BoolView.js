/*
This shows a line chart of the booleans
*/
import {addView} from '../datatypes';
import {generateLineChart} from './components/LineChart';
import {generateDropdownLineChart, generateTimeOptions} from './components/DropdownLineChart';
import dropdownTransformDisplay from './components/dropdownTransformDisplay';

const BoolView = [
    {
        ...generateLineChart(),
        key: "lineView",
        title: "Plot",
        subtitle: ""
    }
];

function showBoolView(context) {
    if (context.data.length > 1) {

        // We now check if the data is booleans
        if (typeof(context.data[0].d) === "boolean" && typeof(context.data[context.data.length - 1].d) === "boolean") {
            return BoolView;
        }

    }

    return null;
}

addView(showBoolView);
