import React, {Component, PropTypes} from 'react';
import DataTable from './DataTable';
import CSVView from './CSVView';

import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import {addView} from '../../datatypes';

class TableView extends Component {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        state: PropTypes.object.isRequired
    }

    render() {
        if (this.props.state.csv === undefined || this.props.state.csv === false) {
            return (<DataTable {...this.props}/>)
        }
        return (<CSVView {...this.props}/>);
    }
}

function getIcons(context) {
    let resultarray = [];

    if (context.state.csv) {
        resultarray.push((
            <IconButton key="csv" onTouchTap={() => context.setState({csv: false})} tooltip="view data table">
                <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                    view_list
                </FontIcon>
            </IconButton>
        ));
    } else {
        resultarray.push((
            <IconButton key="csv" onTouchTap={() => context.setState({csv: true})} tooltip="view data as csv">
                <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                    view_headline
                </FontIcon>
            </IconButton>
        ));

        if (context.data.length > 20) {
            if (context.state.tableExpanded) {
                resultarray.push((
                    <IconButton key="tableExpanded" onTouchTap={() => context.setState({tableExpanded: false})} tooltip="hide older datapoints">
                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                            keyboard_arrow_up
                        </FontIcon>
                    </IconButton>
                ));
            } else {
                resultarray.push((
                    <IconButton key="tableExpanded" onTouchTap={() => context.setState({tableExpanded: true})} tooltip="view hidden datapoints">
                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                            keyboard_arrow_down
                        </FontIcon>
                    </IconButton>
                ));
            }
        }
    }

    return resultarray;
}

const tableView = {
    key: "tableView",
    component: TableView,
    width: "expandable-half",
    initialState: {
        csv: false,
        tableExpanded: false
    },
    title: (context) => {
        if (context.state.csv === undefined || context.state.csv === false) {
            return "Data Table";
        }
        return "Data CSV";
    },
    subtitle: "",
    icons: getIcons
};

// showTable determines whether the table should be shown for the given context
function showTable(context) {
    if (context.data.length > 0) {
        return [tableView];
    }
    return null;
}

addView(showTable);
