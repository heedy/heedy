/*
The CSVView displays the currently queried data as a CSV text, that can be copied to clipboard and imported
into excel and such
*/

import React, {Component, PropTypes} from 'react';
import moment from 'moment';

import DataUpdater from '../components/DataUpdater';

import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import CodeMirror from 'react-codemirror';

class CSVView extends DataUpdater {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        state: PropTypes.object.isRequired,
        setState: PropTypes.func.isRequired
    }

    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        let dataset = ""

        for (let i = 0; i < d.length; i++) {
            dataset += moment.unix(d[i].t).format() + ", " + JSON.stringify(d[i].d) + "\n";
        }

        return dataset;
    }

    render() {
        return (<CodeMirror value={this.state.data} options={{
            lineWrapping: true,
            readOnly: true
        }}/>);
    }

}
export default CSVView;
