/*
This shows a line chart of the data given. It is assumed that the default transform is already set by the default state (transform property).
*/

import React, { Component, PropTypes } from "react";
import BarChart, { getBarChartIcons } from "./BarChart";
import DropDownMenu from "material-ui/DropDownMenu";
import MenuItem from "material-ui/MenuItem";

import dropdownTransformDisplay from "./dropdownTransformDisplay";

class DropdownBarChart extends Component {
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
        <BarChart {...this.props} transform={state.transform} />
      </div>
    );
  }
}
export default DropdownBarChart;

// generate creates a new view that shows the dropdown bar chart, and auto-updates
export function generateDropdownBarChart(
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
    icons: getBarChartIcons,
    component: React.createClass({
      render: function() {
        return <DropdownBarChart {...this.props} options={options} />;
      }
    })
  };
}
