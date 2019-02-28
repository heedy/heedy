/*
The TimeChooser is a component that can be used as the dropdown of an input to permit
inserting datapoints at a specific time.

The main difficulty here is that the inserts will always succeed, even if the timestamp is
before an existing datapoint. This is because inserts are restamp default.
*/

import React, { Component } from "react";
import PropTypes from "prop-types";
import Toggle from "material-ui/Toggle";

import "bootstrap-daterangepicker/daterangepicker.css";
import DateRangePicker from "react-bootstrap-daterangepicker";
import moment from "moment";

// get the timestamp from the current state
export function getTimestamp(state) {
  if (
    state.customtimestamp === undefined ||
    state.customtimestamp == false ||
    state.timestamp === undefined
  ) {
    return moment();
  }

  return state.timestamp;
}

class TimeChooser extends Component {
  static propTypes = {
    state: PropTypes.object.isRequired,
    setState: PropTypes.func.isRequired
  };

  render() {
    let state = this.props.state;

    let customtimestamp =
      state.customtimestamp !== undefined && state.customtimestamp == true;

    let timestamp = state.timestamp !== undefined ? state.timestamp : moment();

    return (
      <div>
        <Toggle
          label="Custom Timestamp"
          labelPosition="right"
          toggled={customtimestamp}
          onToggle={(v, d) => {
            this.props.setState({ customtimestamp: d });
          }}
          trackStyle={{
            backgroundColor: "#ff9d9d"
          }}
        />
        {" "}
        {!customtimestamp
          ? null
          : <div>
              <DateRangePicker
                startDate={state.timestamp}
                singleDatePicker={true}
                opens="left"
                timePicker={true}
                onEvent={(e, picker) =>
                  this.props.setState({ timestamp: picker.startDate })}
              >
                <div
                  id="reportrange"
                  className="selected-date-range-btn"
                  style={{
                    background: "#fff",
                    cursor: "pointer",
                    padding: "5px 10px",
                    border: "1px solid #ccc",
                    width: "100%",
                    textAlign: "center"
                  }}
                >
                  <i className="glyphicon glyphicon-calendar fa fa-calendar pull-right" />&nbsp;

                  <span>{timestamp.format("YYYY-MM-DD hh:mm:ss a")}</span>
                </div>
              </DateRangePicker>
              <p>
                Make sure your timestamp is greater than all existing
                datapoints!
              </p>
            </div>}
      </div>
    );
  }
}

export default TimeChooser;
