/*
  This is the window shown on dropdown for certain views that allows showing a quick description
  of the view's backend
*/

import React from "react";
import PropTypes from "prop-types";
import createClass from "create-react-class";

import TransformInput from "../../../components/TransformInput";

import { app } from "../../../util";
import { setSearchText } from "../../../actions";

export default function generateDropdownTransformDisplay(
  description,
  transform
) {
  return createClass({
    render: function() {
      let tf = transform;
      if (this.props.state.transform !== undefined) {
        tf = this.props.state.transform;
      }
      let desc = description;
      if (this.props.state.description !== undefined) {
        desc = this.props.state.description;
      }
      return (
        <div>
          <p>{desc}</p>
          <h4
            style={{
              paddingTop: "10px"
            }}
          >
            Transform
          </h4>
          <p>This is the transform used to generate this visualization:</p>
          <TransformInput
            transform={tf}
            onChange={txt => null}
            onClick={() => app.dispatch(setSearchText(tf))}
          />
        </div>
      );
    }
  });
}
