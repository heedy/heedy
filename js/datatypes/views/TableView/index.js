import DataTable from './DataTable';
import {addView} from '../../datatypes';

const tableView = {
    key: "tableView",
    component: DataTable,
    width: "expandable-half",
    initialState: {},
    title: "Data Table",
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
