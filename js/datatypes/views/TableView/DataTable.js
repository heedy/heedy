/*
The DataTable displays a table of the currently queried data. It initially only shows up to 20 datapoints (10 first, 10 last),
but can be set to the "expanded" state, where it shows the entire dataset
*/

import React, {Component, PropTypes} from 'react';
import moment from 'moment';

import DataUpdater from '../components/DataUpdater';

class DataTable extends DataUpdater {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        state: PropTypes.object.isRequired,
        setState: PropTypes.func.isRequired
    }

    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        let dataset = new Array(d.length);

        for (let i = 0; i < d.length; i++) {
            dataset[i] = {
                key: JSON.stringify(d[i]),
                t: moment.unix(d[i].t).calendar(),
                d: JSON.stringify(d[i].d)
            };
        }

        return dataset;
    }

    render() {
        let expanded = this.props.state.tableExpanded;

        let data = this.data;
        let expandedText = null;

        // If the table is not expanded, show only the last 10 if there are more than 20
        if (!expanded && data.length > 20) {

            expandedText = (
                <div style={{
                    width: "100%",
                    textAlign: "center"
                }}>
                    <a class="pull-center" style={{
                        cursor: "pointer"
                    }} onClick={() => this.props.setState({tableExpanded: true})}>
                        Show {(data.length - 10).toString() + " "}
                        hidden datapoints
                    </a>
                </div>
            );
            data = data.slice(data.length - 10, data.length);
        }
        return (
            <div>
                <table className="table table-striped" style={{
                    width: "100%",
                    overflow: "auto"
                }}>
                    <thead>
                        <tr>
                            <th>Timestamp</th>
                            <th>Data</th>
                        </tr>
                    </thead>
                    <tbody>
                        {data.map((s) => {
                            return (
                                <tr key={s.key}>
                                    <td>{s.t}</td>
                                    <td>
                                        {s.d}
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
                {expandedText}
            </div>
        );
    }

}
export default DataTable;
