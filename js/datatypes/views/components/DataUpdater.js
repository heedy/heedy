import React, { Component, PropTypes } from 'react';

class DataUpdater extends Component {

    // Sets up the transform for use in the component
    initTransform(t) {
        console.log("initTransform");
        if (t !== undefined && t !== "") {
            try {

                this.transform = this.props.pipescript.Script(t);
                return;
            } catch (e) {
                console.error("TRANSFORM ERROR: ", t, e.toString());
            }
        }
        this.transform = null;

    }

    // Returns the data transformed if there is a transform
    // in the props, and the original data if it is not
    dataTransform(data) {
        console.log("runTransform");
        if (this.transform != null) {
            try {
                return this.transform.Transform(data);
            } catch (e) {
                console.error("DATA ERROR: ", t, e.toString());
            }

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
            console.log("changing");
            if (this.props.transform !== p.transform) {
                this.initTransform(p.transform);
            }
            this.data = this.transformDataset(this.dataTransform(p.data));
        }
    }
}

//export default DataUpdater;
export default DataUpdater;
