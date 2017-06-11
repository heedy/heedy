import React, { Component } from "react";
import PropTypes from "prop-types";

class DataUpdater extends Component {
  // Sets up the transform for use in the component
  initTransform(t) {
    if (t !== undefined && t !== "") {
      try {
        this.transform = this.props.pipescript.Script(t);
        return;
      } catch (e) {
        console.error("TRANSFORM ERROR: ", t, e.toString());
      }
    }
    this.transform = null;
  }

  // Returns the data transformed if there is a transform
  // in the props, and the original data if it is not
  dataTransform(data) {
    if (this.transform != null) {
      try {
        return this.transform.Transform(data);
      } catch (e) {
        console.error("DATA ERROR: ", data, e.toString());
      }
    }
    return data;
  }

  // Generate the component's initial state
  componentWillMount() {
    this.initTransform(this.props.transform);
    this.tdata = this.dataTransform(this.props.data);
    this.data = this.transformDataset(this.tdata, this.props.state);
  }

  // Each time either the data or the transform changes, reload
  componentWillReceiveProps(p) {
    // We only perform the dataset transform operation if the dataset
    // was modified
    if (
      p.data !== this.props.data ||
      this.props.transform !== p.transform ||
      this.props.state !== p.state
    ) {
      if (this.props.transform !== p.transform) {
        this.initTransform(p.transform);
      }
      if (p.data !== this.props.data || this.props.transform !== p.transform) {
        this.tdata = this.dataTransform(p.data);
      }

      this.data = this.transformDataset(this.tdata, p.state);
    }
  }

  shouldComponentUpdate(p, s) {
    if (
      p.data !== this.props.data ||
      this.props.transform !== p.transform ||
      this.props.state !== p.state ||
      s !== this.state
    ) {
      return true;
    }
    return false;
  }
}

//export default DataUpdater;
export default DataUpdater;
