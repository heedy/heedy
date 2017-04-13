/*
This shows a bar chart if the data was manually aggregated into a single-datapoint object with all-number
fields
*/

import { addView } from "../datatypes";
import { generateBarChart } from "./components/BarChart";

const SingleObjectView = [
  {
    ...generateBarChart(),
    key: "objectView",
    title: "Bar/Pie Chart",
    subtitle: ""
  }
];

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
  return !isNaN(parseFloat(n)) && isFinite(n);
}

function showSingleObjectView(context) {
  if (context.data.length != 1 || context.pipescript === null) {
    return null;
  }

  // We check if the object has keys
  let d = context.data[0].d;
  let keys = Object.keys(d);
  if (keys.length > 1) {
    for (let i = 0; i < keys.length; i++) {
      if (!isNumeric(d[keys[i]])) {
        return null;
      }
    }
    return SingleObjectView;
  }
  return null;
}

addView(showSingleObjectView);
