/*
This view shows a histogram of numeric values, binned by a custom bin size
*/

import { addView } from '../datatypes';
import { generateDropdownBarChart } from './components/DropdownBarChart';

import { numeric } from './typecheck';

function HistView(key) {
    let pretransform = (key !== "" ? "$('" + key + "') | " : "");
    return [
        {
            ...generateDropdownBarChart("This view creates a histogram of your numeric values, with each bar being the bucket width.", [
                {
                    name: "Bar Size: 5",
                    transform: pretransform + "map(bucket(5),count)"
                }, {
                    name: "Bar Size: 10",
                    transform: pretransform + "map(bucket(10),count)"
                }, {
                    name: "Bar Size: 20",
                    transform: pretransform + "map(bucket(20),count)"
                }, {
                    name: "Bar Size: 50",
                    transform: pretransform + "map(bucket(50),count)"
                }, {
                    name: "Bar Size: 100",
                    transform: pretransform + "map(bucket(100),count)"
                }, {
                    name: "Bar Size: 500",
                    transform: pretransform + "map(bucket(500),count)"
                }, {
                    name: "Bar Size: 1000",
                    transform: pretransform + "map(bucket(1000),count)"
                }
            ], 1),
            key: "histView",
            title: "Histogram",
            subtitle: ""
        }
    ];
}

function showHistogramView(context) {
    let d = context.data;
    if (d.length < 7 || context.pipescript === null) {
        return null;
    }

    let n = numeric(context.data);

    if (n !== null && !n.allbool) {
        return HistView(n.key);
    }

    return null;
}

addView(showHistogramView);
