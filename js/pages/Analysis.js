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
        )]}>
        <TimePicker state={state} setState={actions.setState} />

        <div className="row">
            <div className="col-md-4">
                <h5 style={{
                    paddingTop: "10px",
                    fontWeight: "bold"
                }}>Reference Stream (X)</h5>
                <TextField id={"X_dataset_text_field"} hintText="user/device/stream" style={{ width: "100%" }}
                    value={state.stream} onChange={(e) => actions.setState({ stream: e.target.value })} />
            </div>
            <div className="col-md-8">
                <h5 style={{
                    paddingTop: "10px"
                }}>Transform</h5>
                <TransformInput transform={state.transform} onChange={(txt) => actions.setState({ transform: txt })} />
            </div>
        </div>
        <DatasetStream name="Y" state={state.dataset.y} setState={(v) => actions.setDatasetState("y", v)} />
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
const Render = ({ state, actions }) => (
    <DataView data={state.data}>
        <AnalysisQuery state={state} actions={actions} />
        {state.loading ? (<div className="col-lg-12"><Loading /></div>) : null}
    </DataView>
);

export default connect(
    (state) => ({ state: state.pages.analysis }),
    (dispatch) => ({ actions: bindActionCreators(Actions, dispatch) })
)(Render);