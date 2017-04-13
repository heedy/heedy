/*
This shows a bar chart if we deem the data to be amenable to bar-chart-plotting.
The exact conditions are in the function showBarChart below
*/

import { addView } from "../datatypes";
import { generateBarChart } from "./components/BarChart";
import { categorical, objectvalues } from "./typecheck";

const BarView = [
  {
    ...generateBarChart(
      "map($,count) | top(20)",
      "Counts the occurences of the given values"
    ),
    key: "barView",
    title: "Value Counts",
    subtitle: ""
  }
];

function meanBarChart(key1, key2) {
  return [
    {
      ...generateBarChart(
        `map($('${key1}'),$('${key2}'):mean)`,
        "Finds the mean of " + key2 + " for each " + key1 + " value.",
        key1,
        "Mean of " + key2
      ),
      key: "meanBarChart",
      title: `Mean of ${key2} for all ${key1}`,
      subtitle: ""
    }
  ];
}

function showBarChart(context) {
  if (context.data.length < 5 || context.pipescript === null) {
    return null;
  }
  if (categorical(context.data) !== null) {
    return BarView;
  }
  let o = objectvalues(context.data);
  if (o !== null && Object.keys(o).length == 2) {
    let k = Object.keys(o);
    if (
      o[k[0]].categorical !== null &&
      o[k[1]].numeric !== null &&
      o[k[0]].categorical.categories < 50
    ) {
      return meanBarChart(k[0], k[1]);
    }
    if (
      o[k[1]].categorical !== null &&
      o[k[0]].numeric !== null &&
      o[k[1]].categorical.categories < 50
    ) {
      return meanBarChart(k[1], k[0]);
    }
  }
  return null;
}

addView(showBarChart);
