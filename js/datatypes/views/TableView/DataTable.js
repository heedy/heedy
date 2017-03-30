/*
The DataTable displays a table of the currently queried data. It initially only shows up to 20 datapoints (10 first, 10 last),
but can be set to the "expanded" state, where it shows the entire dataset
*/

import React, { Component, PropTypes } from 'react';
import moment from 'moment';

import DataUpdater from '../components/DataUpdater';

class DataTable extends DataUpdater {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        state: PropTypes.object.isRequired,
        setState: PropTypes.func.isRequired
    }

    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d, state) {
        let dataset = new Array(d.length);
        let istime = state.timeon;
        if (d.length > 0) {
            // In order to show columns in the data table, we first check if the datapoints are objects...
            // If they are, then we generate the table so that the object is the columns
            if (d.length == 1 && d[0].d !== null && typeof d[0].d === 'object' && Object.keys(d[0].d).length > 4) {
                // It is a single datapoint. We render it as a special data table of key-values
                let t = d[0].t;
                d = d[0].d;
                let keys = Object.keys(d);
                dataset = new Array(keys.length);

                for (let i = 0; i < keys.length; i++) {
                    dataset[i] = {
                        key: keys[i],
                        t: "",
                        d: {
                            Key: keys[i],
                            Value: istime ? moment.duration(d[keys[i]], 'seconds').humanize() : d[keys[i]]
                        }
                    };
                }
                dataset[0].t = moment.unix(t).calendar();
            } else if (d.length == 1 && d[0].d !== null && typeof d[0].d !== 'object') {
                // This is a single datapoint with a single value. Show an alternate view
                // That highlights the value
                dataset[0] = {
                    key: "singleDP",
                    t: moment.unix(d[0].t).calendar(),
                    d: istime ? moment.duration(d[0].d, 'seconds').humanize() : JSON.stringify(d[0].d, undefined, 2)
                };
            }
            else if (d[0].d !== null && typeof d[0].d === 'object' && Object.keys(d[0].d).length < 10) {
                for (let i = 0; i < d.length; i++) {
                    let data = {};
                    Object.keys(d[i].d).map((key) => {
                        data[key.capitalizeFirstLetter()] = istime ? moment.duration(d[i].d[key], 'seconds').humanize() : JSON.stringify(d[i].d[key], undefined, 2);
                    });
                    dataset[i] = {
                        key: JSON.stringify(d[i]),
                        t: moment.unix(d[i].t).calendar(),
                        d: data
                    };
                }
            } else {
                for (let i = 0; i < d.length; i++) {
                    dataset[i] = {
                        key: JSON.stringify(d[i]),
                        t: moment.unix(d[i].t).calendar(),
                        d: {
                            Data: istime ? moment.duration(d[i].d, 'seconds').humanize() : JSON.stringify(d[i].d, undefined, 2)
                        }
                    };
                }
            }
        }




        return dataset;
    }

    render() {
        let expanded = this.props.state.tableExpanded;

        let data = this.data;
        let expandedText = null;

        if (data.length == 1 && data[0].key === "singleDP") {
            return (
                <div style={{ textAlign: "center" }}>
                    <h6>{data[0].t}</h6>
                    {data[0].d.length <= 15 ? (<h1>{data[0].d}</h1>) : (<h3>{data[0].d}</h3>)}
                </div>
            );
        }

        // If the table is not expanded, show only the last 5 if there are more than 10
        if (!expanded && data.length > 10) {

            expandedText = (
                <div style={{
                    width: "100%",
                    textAlign: "center"
                }}>
                    <a className="pull-center" style={{
                        cursor: "pointer"
                    }} onClick={() => this.props.setState({ tableExpanded: true })}>
                        Show {(data.length - 5).toString() + " "}
                        hidden datapoints
                    </a>
                </div>
            );
            data = data.slice(data.length - 5, data.length);
        }
        return (
            <div style={{ maxHeight: 600, overflowY: "auto" }}>
                <table className="table table-striped" style={{
                    width: "100%",
                    overflow: "auto"
                }}>
                    <thead>
                        <tr>
                            <th>Timestamp</th>
                            {data.length === 0 ? (<th>Data</th>) : Object.keys(data[0].d).map((k) => (<th key={k}>{k}</th>))}
                        </tr>
                    </thead>
                    <tbody>
                        {data.map((s) => {
                            return (
                                <tr key={s.key}>
                                    <td>{s.t}</td>
                                    {Object.keys(s.d).map((k) => (<td key={k}>{s.d[k]}</td>))}
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
