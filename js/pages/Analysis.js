import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { TimePicker } from '../components/QueryRange';
import TransformInput from '../components/TransformInput';
import Loading from '../components/Loading';
import ExpandableCard from '../components/ExpandableCard';

import FlatButton from 'material-ui/FlatButton';
import TextField from 'material-ui/TextField';
import SelectField from 'material-ui/SelectField';
import MenuItem from 'material-ui/MenuItem';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import DataView from '../components/DataView';

import * as Actions from '../actions/analysis';
import SearchCard from '../components/SearchCard';
import { setSearchSubmit, setSearchState } from '../actions';

const Interpolator = ({ value, onChange }) => (
    <SelectField
        value={value}
        onChange={(e, i, v) => onChange(v)}
        style={{ width: "100%" }}
    >
        <MenuItem value={"closest"} primaryText="closest" />
        <MenuItem value={"before"} primaryText="before" />
        <MenuItem value={"after"} primaryText="after" />
        <MenuItem value={"count"} primaryText="count" />
        <MenuItem value={"mean"} primaryText="mean" />
        <MenuItem value={"sum"} primaryText="sum" />
    </SelectField>
)

const DatasetStream = ({ name, state, setState }) => (
    <div className="row">
        <div className="col-lg-4">
            <h5 style={{
                paddingTop: "10px",
                fontWeight: "bold"
            }}>Interpolated Stream ({name})</h5>
            <TextField style={{ width: "100%" }} id={name + "_dataset_text_field"} hintText="user/device/stream"
                value={state.stream} onChange={(e) => setState({ stream: e.target.value })} />
        </div>
        <div className="col-lg-5">
            <h5 style={{
                paddingTop: "10px"
            }}>Transform</h5>
            <TransformInput transform={state.transform} onChange={(txt) => setState({ transform: txt })} />
        </div>
        <div className="col-lg-3">
            <h5 style={{
                paddingTop: "10px"
            }}>Interpolator</h5>
            <Interpolator value={state.interpolator} onChange={(v) => setState({ interpolator: v })} />
        </div>
    </div>
);

const XDataset = ({ state, actions }) => (
    <div className="row">
        <div className="col-md-4">
            <h5 style={{
                paddingTop: "10px",
                fontWeight: "bold"
            }}>Reference Stream (x)</h5>
            <TextField id={"X_dataset_text_field"} hintText="user/device/stream" style={{ width: "100%" }}
                value={state.stream} onChange={(e) => actions.setState({ stream: e.target.value })} />
        </div>
        <div className="col-md-8">
            <h5 style={{
                paddingTop: "10px"
            }}>Transform<a className="pull-right" style={{
                cursor: "pointer"
            }} onClick={() => actions.setState({ xdataset: false })}>Switch to T-Dataset</a></h5>
            <TransformInput transform={state.transform} onChange={(txt) => actions.setState({ transform: txt })} />
        </div>
    </div>
);

const TDataset = ({ state, actions }) => (
    <div className="row" style={{ paddingTop: "10px" }}>
        <div className="col-md-2">
            <h5 style={{
                paddingTop: "10px",
                fontWeight: "bold"
            }}>Time Delta</h5>
        </div>
        <div className="col-md-4">
            <TextField id={"DT"} style={{ width: "100%" }}
                value={state.dt} onChange={(e) => actions.setState({ dt: e.target.value })} />
        </div>
        <div className="col-md-3">
            <SelectField
                value={state.dt}
                onChange={(e, i, v) => actions.setState({ dt: v })}
                style={{ width: "100%" }}
            >
                <MenuItem value={"1"} primaryText="1 second" />
                <MenuItem value={"60"} primaryText="1 minute" />
                <MenuItem value={"1800"} primaryText="30 minutes" />
                <MenuItem value={"3600"} primaryText="1 hour" />
                <MenuItem value={"21600"} primaryText="6 hours" />
                <MenuItem value={"43200"} primaryText="12 hours" />
                <MenuItem value={"86400"} primaryText="1 day" />
                <MenuItem value={"604800"} primaryText="1 week" />
                <MenuItem value={"1296000"} primaryText="15 days" />
                <MenuItem value={"2592000"} primaryText="30 days" />
                <MenuItem value={"31536000"} primaryText="1 year" />
            </SelectField>
        </div>
        <div className="col-md-3">
            <h5><a className="pull-right" style={{
                cursor: "pointer"
            }} onClick={() => actions.setState({ xdataset: true })}>Switch to X-Dataset</a></h5>
        </div>
    </div>
);


/**
 * Component to display the analysis form, from which you can generate a query for analysis,
 * and run the query
 * @param {state, action} the state and action objects for component control
 */
const AnalysisQuery = ({ state, actions }) => (
    <ExpandableCard state={state} width="full" setState={actions.setState}
        title="Analyze Data Streams" subtitle="Correlate datasets involving multiple streams"
        icons={[(
            <IconButton key="clearanalysis" onTouchTap={actions.clear} tooltip="Clear Analysis">
                <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                    clear_all
                </FontIcon>
            </IconButton>
        ), (
            <IconButton key="pythoncode" onTouchTap={actions.showPython} tooltip="Show Python Code">
                <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                    code
                </FontIcon>
            </IconButton>
        )]}>
        <TimePicker state={state} setState={actions.setState} />
        {state.xdataset ? (<XDataset state={state} actions={actions} />) : (<TDataset state={state} actions={actions} />)}
        {Object.keys(state.dataset).map((k) => (<DatasetStream key={k} name={k} state={state.dataset[k]} setState={(v) => actions.setDatasetState(k, v)} />))}
        <div>
            <IconButton tooltip="Add Stream" onTouchTap={actions.addDatasetStream}>
                <FontIcon className="material-icons" color={Object.keys(state.dataset).length >= 10 ? "grey" : "black"}>add</FontIcon>
            </IconButton>
            <IconButton tooltip="Remove Stream" onTouchTap={actions.removeDatasetStream}>
                <FontIcon className="material-icons" color={Object.keys(state.dataset).length == 1 ? "grey" : "black"}>remove</FontIcon>
            </IconButton>
        </div>
        <h5 style={{
            paddingTop: "10px"
        }}>Server-Side Transform to Run After Generating Dataset</h5>
        <TransformInput transform={state.posttransform} onChange={(txt) => actions.setState({ posttransform: txt })} />

        <FlatButton style={{
            float: "right"
        }} primary={true} label="Generate Dataset" onTouchTap={() => actions.query()} /> {state.error !== null
            ? (
                <p style={{
                    paddingTop: "10px",
                    color: "red"
                }}>{state.error}</p>
            )
            : (
                <p style={{
                    paddingTop: "10px"
                }}>Learn about datasets <a href="https://connectordb.io/docs/datasets/">{" "}here</a>{", "}and transforms
                            <a href="https://connectordb.io/docs/pipescript/">{" "}here.</a>
                </p>
            )}
    </ExpandableCard>
)

/**
 * Render controls the display of the entire analysis page. 
 * The associated reducer and saga are in their corresponding directories
 * 
 */
const Render = ({ state, actions, transformError, clearTransform }) => (
    <div>
        {state.search.submitted != null && state.search.submitted != ""
            ? (<SearchCard title={state.search.submitted} subtitle={"Transform applied to data"} onClose={clearTransform} />)
            : null}
        <DataView data={state.data} transform={state.search.submitted} transformError={transformError} >
            <AnalysisQuery state={state} actions={actions} />
            {state.loading ? (<div className="col-lg-12"><Loading /></div>) : null}
        </DataView>
    </div>
);

export default connect(
    (state) => ({ state: state.pages.analysis }),
    (dispatch) => ({
        actions: bindActionCreators(Actions, dispatch),
        clearTransform: () => dispatch(setSearchSubmit("")),
        transformError: (txt) => dispatch(setSearchState({ error: txt }))
    }),
)(Render);