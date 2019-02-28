/*
This shows a line chart of the booleans
*/
import { addView } from "../datatypes";
import { generateLineChart } from "./components/LineChart";
import {
  generateDropdownLineChart,
  generateTimeOptions
} from "./components/DropdownLineChart";
import dropdownTransformDisplay from "./components/dropdownTransformDisplay";

import { numeric } from "./typecheck";

const BoolView = [
  {
    ...generateLineChart(),
    key: "boolView",
    title: "Boolean View",
    subtitle: ""
  }
];

function showBoolView(context) {
  if (context.data.length > 1) {
    // We now check if the data is booleans
    let n = numeric(context.data);

    if (n !== null && n.allbool) {
      return BoolView;
    }
  }

  return null;
}

addView(showBoolView);
