/*
 This is the line chart component
*/

import React, {Component, PropTypes} from 'react';
import {Line} from 'react-chartjs';
import moment from 'moment';

// generateDatasetFromData generates the dataset in the format used by chartjs
// from an array of data. ConnectorDB uses the format:
//  {t: unix floating point timestamp, d: data}
// but chartjs uses:
// {x: momentjs timestamp,y: data}
// so we convert one to the other
function generateDatasetFromData(name, d) {
    let dataset = new Array(d.length);

    for (let i = 0; i < d.length; i++) {
        dataset[i] = {
            x: moment.unix(d[i].t),
            y: d[i].d
        }
    }

    return [
        {
            label: name,
            data: dataset,
            lineTension: 0
        }
    ];
}

// generateDatasetFromObject is the same idea as generateDatasetFromData.
// The difference between the two is that this function expects each datapoint
// to be an object. This means that the data is actually multiple "series", which
// can be shown on a legend.
function generateDatasetFromObject(d) {
    var resultmap = {};

    // Loop through the array generating the datasets as we go
    for (let i = 0; i < d.length; i++) {
        let t = moment.unix(d[i].t);
        Object.keys(d[i]).forEach((key) => {
            if (resultmap[key] === undefined) {
                resultmap[key] = [];
            }
            resultmap[key].push({x: t, y: d[i].d[key]});
        });
    }

    // We now have an object with the datapoints as arrays. We now split it into one
    // large array of datasets
    var result = [];
    Object.keys(resultmap).forEach((key) => {
        result.push({label: key, data: resultmap[key], lineTension: 0});
    });

    return result;
}

class LineChart extends Component {}
