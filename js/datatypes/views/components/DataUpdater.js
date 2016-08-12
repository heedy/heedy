import React, {Component, PropTypes} from 'react';

class DataUpdater extends Component {

    // Sets up the transform for use in the component
    initTransform(t) {
        if (t !== undefined && t !== "") {
            this.transform = this.props.pipescript.Script(t);
        } else {
            this.transform = null;
        }
    }

    // Returns the data transformed if there is a transform
    // in the props, and the original data if it is not
    dataTransform(data) {
        if (this.transform != null) {
            return this.transform.Transform(data);
        }
        return data;
    }

    // Generate the component's initial state
    componentWillMount() {
        this.initTransform(this.props.transform);
        this.data = this.transformDataset(this.dataTransform(this.props.data));
    }

    // Each time either the data or the transform changes, reload
    componentWillReceiveProps(p) {
        // We only perform the dataset transform operation if the dataset
        // was modified
        if (p.data !== this.props.data || this.props.transform !== p.transform) {
            if (this.props.transform !== p.transform) {
                this.initTransform(p.transform);
            }
            this.data = this.transformDataset(this.dataTransform(p.data));
        }
    }
}

//export default DataUpdater;
export default DataUpdater;
