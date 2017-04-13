import React, { Component, PropTypes } from "react";
import DataTable from "./DataTable";
import CSVView from "./CSVView";
import DataUpdater from "../components/DataUpdater";
import dropdownTransformDisplay from "../components/dropdownTransformDisplay";

import FontIcon from "material-ui/FontIcon";
import IconButton from "material-ui/IconButton";

import { addView } from "../../datatypes";
import { numeric } from "../typecheck";

class TableView extends Component {
  static propTypes = {
    data: PropTypes.arrayOf(PropTypes.object).isRequired,
    state: PropTypes.object.isRequired
  };

  render() {
    if (this.props.state.csv === undefined || this.props.state.csv === false) {
      return <DataTable {...this.props} />;
    }
    return <CSVView {...this.props} />;
  }
}

function getIcons(context) {
  let resultarray = [];

  if (context.state.csv) {
    resultarray.push(
      <IconButton
        key="csv"
        onTouchTap={() => context.setState({ csv: false })}
        tooltip="view data table"
      >
        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
          view_list
        </FontIcon>
      </IconButton>
    );
  } else {
    resultarray.push(
      <IconButton
        key="csv"
        onTouchTap={() => context.setState({ csv: true })}
        tooltip="view data as csv"
      >
        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
          view_headline
        </FontIcon>
      </IconButton>
    );

    if (context.state.allowtime) {
      if (context.state.timeon) {
        resultarray.push(
          <IconButton
            key="timer"
            onTouchTap={() => context.setState({ timeon: false })}
            tooltip="disble pretty printing duration"
          >
            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
              alarm_off
            </FontIcon>
          </IconButton>
        );
      } else {
        resultarray.push(
          <IconButton
            key="timer"
            onTouchTap={() => context.setState({ timeon: true })}
            tooltip="view data as time duration"
          >
            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
              access_time
            </FontIcon>
          </IconButton>
        );
      }
    }

    if (context.data.length > 20) {
      if (context.state.tableExpanded) {
        resultarray.push(
          <IconButton
            key="tableExpanded"
            onTouchTap={() => context.setState({ tableExpanded: false })}
            tooltip="hide older datapoints"
          >
            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
              arrow_drop_up
            </FontIcon>
          </IconButton>
        );
      } else {
        resultarray.push(
          <IconButton
            key="tableExpanded"
            onTouchTap={() => context.setState({ tableExpanded: true })}
            tooltip="view hidden datapoints"
          >
            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
              arrow_drop_down
            </FontIcon>
          </IconButton>
        );
      }
    }
  }

  return resultarray;
}

function getTableView(allowtime) {
  return {
    key: "tableView",
    component: TableView,
    width: "expandable-half",
    initialState: {
      csv: false,
      tableExpanded: false,
      allowtime: allowtime,
      timeon: false
    },
    title: context => {
      if (context.state.csv === undefined || context.state.csv === false) {
        return "Data";
      }
      return "Data CSV";
    },
    subtitle: "",

    icons: getIcons
  };
}

function getTransformedTableView(
  transform,
  key,
  pretext,
  description,
  allowtime,
  istime = false
) {
  return {
    key: key,
    component: p => <TableView {...p} transform={transform} />,
    width: "expandable-half",
    initialState: {
      csv: false,
      tableExpanded: false,
      allowtime: allowtime,
      timeon: istime
    },
    title: context => {
      if (context.state.csv === undefined || context.state.csv === false) {
        return pretext;
      }
      return pretext + " CSV";
    },
    dropdown: dropdownTransformDisplay(description, transform),
    subtitle: "",
    icons: getIcons
  };
}

// showTable determines whether the table should be shown for the given context
function showTable(context) {
  if (context.data.length > 0) {
    let n = numeric(context.data);
    let t = [getTableView(n !== null && !n.allbool)];

    if (n !== null && n.allbool) {
      t.push(
        getTransformedTableView(
          n.key === "" ? "ttrue" : "$('" + n.key + "') | ttrue",
          "ttruetable",
          "Time True",
          "Time that the stream spends in the true state",
          true,
          true
        )
      );
    }
    return t;
  }
  return null;
}

addView(showTable);
