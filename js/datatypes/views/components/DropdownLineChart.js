/*
This shows a line chart of the data given. It is assumed that the default transform is already set by the default state (transform property).
*/

import React, { Component, PropTypes } from "react";
import LineChart from "./LineChart";
import DropDownMenu from "material-ui/DropDownMenu";
import MenuItem from "material-ui/MenuItem";

import dropdownTransformDisplay from "./dropdownTransformDisplay";

class DropdownLineChart extends Component {
  static propTypes = {
    data: PropTypes.arrayOf(PropTypes.object).isRequired,
    state: PropTypes.object.isRequired,
    setState: PropTypes.func.isRequired,
    options: PropTypes.arrayOf(
      PropTypes.shape({
        description: PropTypes.string,
        transform: PropTypes.string.isRequired,
        name: PropTypes.string.isRequired
      })
    )
  };

  render() {
    const { data, state, setState, options } = this.props;
    return (
      <div
        style={{
          textAlign: "center"
        }}
      >
        <DropDownMenu
          value={state.currentOption}
          onChange={(e, i, v) =>
            setState({
              transform: v.transform,
              currentOption: v,
              description: v.description
            })}
        >
          {options.map(function(o) {
            return (
              <MenuItem key={o.transform} value={o} primaryText={o.name} />
            );
          })}
        </DropDownMenu>
        <LineChart {...this.props} transform={state.transform} />
      </div>
    );
  }
}
export default DropdownLineChart;

// generate creates a new view that shows the dropdown line chart, and auto-updates
export function generateDropdownLineChart(
  description,
  options,
  defaultOptionIndex
) {
  return {
    initialState: {
      currentOption: options[defaultOptionIndex],
      transform: options[defaultOptionIndex].transform,
      description: options[defaultOptionIndex].description
    },
    width: "expandable-half",
    dropdown: dropdownTransformDisplay(description, ""),
    component: React.createClass({
      render: function() {
        return <DropdownLineChart {...this.props} options={options} />;
      }
    })
  };
}

export function generateTimeOptions(type, startup, within) {
  // Add a pipe to the startup text, so we can execute the while after it
  if (startup.length > 0) {
    startup = startup + " | ";
  }
  return [
    {
      name: "Hourly " + type,
      transform: startup + "while(hour==next:hour," + within + ")"
    },
    {
      name: "Daily " + type,
      transform: startup + "while(day==next:day," + within + ")"
    },
    {
      name: "Weekly " + type,
      transform: startup + "while(week==next:week," + within + ")"
    },
    {
      name: "Monthly " + type,
      transform: startup + "while(month==next:month," + within + ")"
    },
    {
      name: "Yearly " + type,
      transform: startup + "while(year==next:year," + within + ")"
    }
  ];
}
