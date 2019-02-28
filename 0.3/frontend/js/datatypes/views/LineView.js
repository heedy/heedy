/*
This shows a line chart of the data given
*/

import { addView } from "../datatypes";
import { generateLineChart, LineChart } from "./components/LineChart";
import {
  generateDropdownLineChart,
  generateTimeOptions
} from "./components/DropdownLineChart";
import dropdownTransformDisplay from "./components/dropdownTransformDisplay";

import { numeric, objectvalues } from "./typecheck";

const BasicLineView = {
  ...generateLineChart(),
  key: "lineView",
  title: "Basic Plot",
  subtitle: ""
};

function lineViewGenerator(pretransform = "", within = "") {
  return [
    {
      ...generateLineChart(pretransform),
      key: "lineView",
      title: "Basic Plot",
      subtitle: "",
      dropdown: pretransform !== ""
        ? dropdownTransformDisplay("", pretransform)
        : undefined
    },
    {
      ...generateDropdownLineChart(
        "This view averages the datapoint values over the chosen time period",
        generateTimeOptions("Average", pretransform, within + "mean"),
        1
      ),
      key: "averagedLineView",
      title: "Averaged Values",
      subtitle: ""
    },
    {
      ...generateDropdownLineChart(
        "This view sums the datapoint values over the chosen time period",
        generateTimeOptions("Sum", pretransform, within + "sum"),
        1
      ),
      key: "summedLineView",
      title: "Summed Values",
      subtitle: ""
    }
  ];
}

const LineView = lineViewGenerator("");

function showLineChart(context) {
  if (context.data.length > 1 && context.pipescript !== null) {
    let n = numeric(context.data);
    if (n !== null && !n.allbool) {
      return lineViewGenerator(n.key !== "" ? "$('" + n.key + "')" : "");
    }
    if (n !== null && n.allbool) {
      return lineViewGenerator(
        n.key !== "" ? "$('" + n.key + "') | ttrue" : "ttrue"
      );
    }

    let o = objectvalues(context.data);
    if (o !== null && Object.keys(o).length <= LineChart.objectColors.length) {
      let k = Object.keys(o);
      for (let i = 0; i < k.length; i++) {
        if (o[k[i]].numeric === null) return null;
      }
      return BasicLineView;
    }
  }

  return null;
}

addView(showLineChart);
