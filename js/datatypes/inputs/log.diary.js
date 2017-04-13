import React, { Component, PropTypes } from "react";
import { addInput } from "../datatypes";

import Textarea from "react-textarea-autosize";
import RaisedButton from "material-ui/RaisedButton";

import TimeChooser, { getTimestamp } from "./TimeChooser";

export const diarySchema = {
  type: "string",
  minLength: 1
};

class Diary extends Component {
  static propTypes = {
    state: PropTypes.object,
    path: PropTypes.string.isRequired,
    setState: PropTypes.func,
    insert: PropTypes.func
  };
  render() {
    let value = this.props.state.value;
    if (value === undefined || value == null) value = "";
    return (
      <div>
        <Textarea
          style={{
            width: "100%",
            fontSize: 18,
            borderColor: "#ccc"
          }}
          value={value}
          minRows={4}
          useCacheForDOMMeasurements
          name={this.props.path}
          multiLine={true}
          onChange={e => this.props.setState({ value: e.target.value })}
        /><br />
        <RaisedButton
          primary={true}
          label="Submit"
          onTouchTap={() =>
            this.props.insert(getTimestamp(this.props.state), value)}
        />
      </div>
    );
  }
}

// add the input to the input registry.
addInput("log.diary", {
  width: "expandable-full",
  component: Diary,
  style: {
    textAlign: "center"
  },
  dropdown: TimeChooser
});
