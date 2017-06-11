/*
Loading shows a cirular progress bar that keeps rotating. This is showed while pages are waiting for required information
*/

import React, { Component } from "react";
import PropTypes from "prop-types";
import CircularProgress from "material-ui/CircularProgress";

export default function Loading() {
  return (
    <div
      style={{
        textAlign: "center",
        paddingTop: 200
      }}
    >
      <CircularProgress />
    </div>
  );
}
