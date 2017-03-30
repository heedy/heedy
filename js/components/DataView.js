/**
 * The DataView represents a view of a dataset. This component renders all plotting
 * and visualization that is available to it. It is given the data, and some optional
 * information about the data. It can also have children, which are rendered before
 * the visualizations. This is used for both the stream and analysis pages, which
 * show plots of data.
 * 
 * The DataView performs extensive caching, since transforming large amounts of datapoints
 * can be very computationally expensive. It only rerenders the visualizations if the relevant
 * properties change.
 * 
 * Furthermore, the DataView handles the state of each visualization. Each visualization's 
 * state goes into the DataView's component state. It might be useful to make the visualization
 * states go into the redux store at some point in the future.
 */

import React, { Component, PropTypes } from 'react';
import { connect } from 'react-redux';

import DataViewCard from '../components/DataViewCard';
import SearchCard from '../components/SearchCard';
import { getViews } from '../datatypes/datatypes';

import { showMessage } from '../actions';

class DataView extends Component {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object),
        datatype: PropTypes.string,
        schema: PropTypes.string,   //
        transform: PropTypes.string, // An optional transform to apply to the data
        transformError: PropTypes.func,

        // These are automatically extracted from the store
        pipescript: PropTypes.object,
        showMessage: PropTypes.func.isRequired
    }

    static defaultProps = {
        data: [],
        schema: "{}",
        datatype: "",
        transform: "",
        transformError: (err) => console.log(err)
    };

    constructor(props) {
        super(props);
        this.state = {};
    }

    // Sets up the transform (if any) so that we're good to go with pipescript
    initTransform(t) {
        if (t !== undefined && t !== "") {
            if (this.props.pipescript != null) {
                try {
                    this.transform = this.props.pipescript.Script(t);
                    return;
                } catch (e) {
                    this.props.transformError(e.toString());
                }

            }
        }
        this.transform = null;
    }

    // Returns the data transformed if there is a transform
    // in the props, and the original data if it is not
    dataTransform(data) {
        if (this.transform != null) {
            try {
                return this.transform.Transform(data);
            } catch (e) {
                this.transform = null;
                this.props.transformError(e.toString());
            }
        }
        return data;
    }

    generateViews(p) {
        this.schema = JSON.parse(p.schema !== "" ? p.schema : "{}");
        // Check which views of data to show
        this.views = getViews({
            data: this.data,    // data was already set earlier
            datatype: p.datatype,
            schema: this.schema,
            pipescript: p.pipescript
        });
        console.log("Showing Views: ", this.views);
    }

    // Generate the component's initial state
    componentWillMount() {
        this.initTransform(this.props.transform);

        this.data = this.dataTransform(this.props.data);

        this.generateViews(this.props);
    }

    // Each time either the data or the transform changes, reload
    componentWillReceiveProps(p) {
        // We only perform the dataset transform operation if the dataset
        // was modified
        if (p.data !== this.props.data || this.props.transform !== p.transform) {
            if (this.props.transform !== p.transform) {
                this.initTransform(p.transform);
            }
            this.data = this.dataTransform(p.data);
        }
        if (p.data !== this.props.data || this.props.transform !== p.transform
            || p.schema !== this.props.schema || p.pipescript !== this.props.pipescript || p.datatype !== this.props.datatype) {
            this.generateViews(p);
        }

    }

    /**
     * Each view can have a state associated with it. The DataView component contains the state
     * for all views, so we need a state updater here.
     * 
     * @param {*} key The name of the view that is using this state
     * @param {*} value An object to update the state
     */
    setViewState(view, value) {
        let newstate = {};
        if (this.state[view.key] === undefined) {
            newstate[view.key] = Object.assign({}, view.initialState, value);
        } else {
            newstate[view.key] = Object.assign({}, this.state[view.key], value);
        }
        console.log("Setting view state for " + view.key, newstate);
        this.setState(newstate);
    }

    getViewState(view) {
        return (this.state[view.key] !== undefined ? this.state[view.key] : view.initialState);
    }

    render() {
        return (
            <div style={{
                marginLeft: "-15px",
                marginRight: "-15px"
            }}>
                {this.props.children}
                {this.views.map((view) => (<DataViewCard key={view.key} view={view} data={this.data}
                    schema={this.schema} datatype={this.props.datatype}
                    state={this.getViewState(view)} setState={(s) => this.setViewState(view, s)}
                    pipescript={this.props.pipescript} msg={this.props.showMessage} />))}
                {this.props.after !== undefined ? this.props.after : null}
            </div>
        );
    }
}

export default connect((state) => ({
    pipescript: state.site.pipescript
}), (dispatch) => ({
    showMessage: (t) => dispatch(showMessage(t))
}))(DataView);