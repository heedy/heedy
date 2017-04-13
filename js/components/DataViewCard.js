/**
  DataViewCard is the card which holds a "data view" - it is given stream details as well as
  the array of datapoints that was queried, and a "view", which is defined in the datatypes folder.
  You can find examples of this card when looking at stream data - all plots and data tables are "views" rendered
  within DataViewCards.

  The DataViewCard sets up the view and displays it within its card. It manages the display size
  of the card, as well as managing the extra dropdown options (if given). This greatly simplifies
  the repeated code used in each view.
**/

import React, { Component, PropTypes } from "react";

import ExpandableCard from "./ExpandableCard";

// Several properties in a view accept both a direct value OR a generator function that
// takes in the current state, and sets the view's value. This function extracts the correct
// value from these properties
function extractValue(value, context) {
  if (typeof value === "function") {
    return value(context);
  }
  return value;
}

class DataViewCard extends Component {
  static propTypes = {
    view: PropTypes.object.isRequired,
    state: PropTypes.object.isRequired,
    setState: PropTypes.func.isRequired,
    schema: PropTypes.object.isRequired,
    datatype: PropTypes.string.isRequired,
    pipescript: PropTypes.object,
    msg: PropTypes.func.isRequired,
    data: PropTypes.arrayOf(PropTypes.object).isRequired
  };

  render() {
    let view = this.props.view;

    let context = {
      schema: this.props.schema,
      datatype: this.props.datatype,
      pipescript: this.props.pipescript,
      state: this.props.state,
      msg: this.props.msg,
      data: this.props.data,
      setState: this.props.setState
    };
    let dropdown = null;
    if (view.dropdown !== undefined) {
      dropdown = <view.dropdown {...context} />;
    }

    return (
      <ExpandableCard
        width={view.width}
        state={this.props.state}
        icons={extractValue(view.icons, context)}
        setState={context.setState}
        dropdown={dropdown}
        title={extractValue(view.title, context)}
        subtitle={extractValue(view.subtitle, context)}
        style={extractValue(view.style, context)}
      >
        <view.component {...context} />
      </ExpandableCard>
    );
  }
}
export default DataViewCard;
