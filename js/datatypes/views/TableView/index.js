import React, {Component, PropTypes} from 'react';
import DataTable from './DataTable';
import CSVView from './CSVView';

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

const tableView = {
    key: "tableView",
    component: TableView,
    width: "expandable-half",
    initialState: {
        csv: false
    },
    title: (state) => {
        if (state.csv === undefined || state.csv === false) {
            return "Data Table";
        }
        return "Data CSV";
    },
    subtitle: ""
};

// showTable determines whether the table should be shown for the given context
function showTable(context) {
    if (context.data.length > 0) {
        return tableView;
    }
    return null;
}

addView(showTable);
