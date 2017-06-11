import React, { Component } from "react";
import PropTypes from "prop-types";

import { addView } from "../datatypes";
import { numeric, objectvalues, dataKeyCompare, getNumber } from "./typecheck";

import math from "mathjs";

const ShowCorrelation = ({ data }) => {
  let o = objectvalues(data);
  let k = Object.keys(o);
  let xdata = data.map(dp => getNumber(dp.d[k[0]]));
  let ydata = data.map(dp => getNumber(dp.d[k[1]]));

  let xmean = math.mean(xdata);
  let ymean = math.mean(ydata);

  let xstd = math.std(xdata);
  let ystd = math.std(ydata);

  let e = 0;
  for (let i = 0; i < xdata.length; i++) {
    e += xdata[i] * ydata[i];
  }

  let correlation =
    (e - xdata.length * xmean * ymean) / ((xdata.length - 1) * xstd * ystd);

  correlation = math.round(correlation, 4);
  return (
    <div style={{ textAlign: "center" }}>
      <h2>{String(correlation)}</h2>
    </div>
  );
};

const CorrelationView = {
  key: "corrView",
  component: ShowCorrelation,
  width: "expandable-half",
  initialState: {},
  title: "Pearson Correlation",
  subtitle: ""
};

function showInfoView(context) {
  let d = context.data;
  if (d.length <= 3) {
    return null;
  }

  let n = numeric(context.data);

  if (n !== null && !n.allbool) {
    return null;
  }

  let o = objectvalues(context.data);
  if (o !== null && Object.keys(o).length === 2) {
    let k = Object.keys(o);
    if (
      o[k[0]].numeric !== null &&
      o[k[1]].numeric !== null &&
      !o[k[0]].numeric.allbool &&
      !o[k[1]].numeric.allbool &&
      o[k[0]].numeric.key == "" &&
      o[k[1]].numeric.key == ""
    ) {
      return CorrelationView;
    }
  }

  return null;
}

addView(showInfoView);
