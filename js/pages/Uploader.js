import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import CodeMirror from 'react-codemirror';

import { Card, CardText, CardHeader } from 'material-ui/Card';
import RaisedButton from 'material-ui/RaisedButton';
import Checkbox from 'material-ui/Checkbox';
import LinearProgress from 'material-ui/LinearProgress';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import TextField from 'material-ui/TextField';

import ExpandableCard from '../components/ExpandableCard';
import AvatarIcon from '../components/AvatarIcon';
import TransformInput from '../components/TransformInput';
import StreamInput from '../components/StreamInput';

import * as Actions from '../actions/uploader';
import { go } from '../actions';

import DataView from '../components/DataView';
import SearchCard from '../components/SearchCard';
import { setSearchSubmit, setSearchState } from '../actions';

// We want to clear the textbox on click, so we need to know the text we can clear
import { UploaderPageInitialState } from '../reducers/uploaderPage';


const Part1Dropdown = ({ state, setState }) => (
    <div className="row">
        <div className="col-sm-12" >
            <p>
                ConnectorDB will try to parse your data as CSV with header or JSON, but sometimes it will need help parsing the timestamp.
            Here, you can optionally set the field name of the timestamp and the <a href="https://momentjs.com/docs/#/parsing/string-format/">timestamp format</a>.
        </p>
        </div>
        <div className="col-sm-6">
            <h5>Timestamp Field</h5>
            <TextField hintText="timestamp" style={{ width: "100%" }}
                value={state.fieldname} onChange={(e) => setState({ fieldname: e.target.value })} />
        </div>
        <div className="col-sm-6">
            <h5>Timestamp Format</h5>
            <TextField hintText="MM-DD-YYYY" style={{ width: "100%" }}
                value={state.timeformat} onChange={(e) => setState({ timeformat: e.target.value })} />
        </div>
    </div>
);

const Part1 = ({ state, actions }) => (
    <ExpandableCard width="expandable-half" state={state.part1} setState={actions.setPart1}
        title={"Step 1"} subtitle={"Add your Data"} avatar={(<AvatarIcon name="paste" iconsrc="material:content_paste" />)}
        icons={[(
            <IconButton key="clearupload" onTouchTap={actions.clear} tooltip="Clear Data">
                <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                    clear_all
                </FontIcon>
            </IconButton>
        )]} dropdown={<Part1Dropdown state={state.part1} setState={actions.setPart1} />}>
        <CodeMirror value={state.part1.rawdata} options={{
            lineWrapping: true,
            mode: "text/plain",
        }} onChange={(txt) => actions.setPart1({ rawdata: txt })} onFocusChange={(f) => (f ? (
            state.part1.rawdata === UploaderPageInitialState.part1.rawdata ? actions.setPart1({ rawdata: "" }) : null) : (
                state.part1.rawdata.trim() === "" ? actions.setPart1({ rawdata: UploaderPageInitialState.part1.rawdata }) : null
            ))} />
    </ExpandableCard>

);

const Part2 = ({ state, actions }) => (
    <ExpandableCard width="expandable-half" state={state.part2} setState={actions.setPart2}
        title={"Step 2"} subtitle={"Check if ConnectorDB can parse the data"} avatar={(<AvatarIcon name="editdd" iconsrc="material:mode_edit" />)} >

        <h5 >Transform{state.part2.error !== ""
            ? (
                <p style={{
                    color: "red",
                    float: "right"
                }}>{state.part2.error}</p>
            )
            : (
                <p style={{
                    float: "right"
                }}>Learn about transforms
                            <a href="https://connectordb.io/docs/pipescript/">{" "}here.</a>
                </p>
            )}</h5>
        <TransformInput transform={state.part2.transform} onChange={(txt) => actions.setPart2({ transform: txt })} />
        <div style={{ textAlign: "center" }}>
            <RaisedButton backgroundColor="#f3f3f3" style={{ marginTop: 10 }} label="Process Data" onTouchTap={actions.process} />
        </div>
    </ExpandableCard >
);

const Part3 = ({ state, actions }) => (
    <ExpandableCard width="expandable-half" state={state.part3} setState={actions.setPart3}
        title={"Step 3"} subtitle={"Upload the Data"} avatar={(<AvatarIcon name="ediy" iconsrc="material:publish" />)}>
        <h5>Stream Name</h5>
        <StreamInput value={state.part3.stream} onChange={(txt) => actions.setPart3({ stream: txt })} />
        <Checkbox label="Create stream if it doesn't exist" checked={state.part3.create} onCheck={(e, c) => actions.setPart3({ create: c })} />
        <Checkbox label="Append data if stream exists" checked={state.part3.overwrite} onCheck={(e, c) => actions.setPart3({ overwrite: c })} />
        <Checkbox label="Ignore datapoints older than data in stream" checked={state.part3.removeolder} onCheck={(e, c) => actions.setPart3({ removeolder: c })} />
        {state.part3.loading ? (<LinearProgress style={{ marginTop: 20, backgroundColor: "#e3e3e3" }} mode={state.part3.percentdone == 0 ? "indeterminate" : "determinate"} value={state.part3.percentdone} />) : (<div style={{ textAlign: "center" }}>
            <RaisedButton backgroundColor="#f3f3f3" style={{ marginTop: 10 }} label="Upload" onTouchTap={actions.upload} />
        </div>)}
        {state.part3.error !== "" ? (<p style={{
            color: "red",
            textAlign: "center",
            paddingTop: 10
        }}>{state.part3.error}</p>) : null}
    </ExpandableCard>
);



const Render = ({ state, actions, go, transformError, clearTransform }) => (
    <div>
        {state.search.submitted != null && state.search.submitted != ""
            ? (<SearchCard title={state.search.submitted} subtitle={"Transform applied to data"} onClose={clearTransform} />)
            : null}
        <DataView data={state.data} transform={state.search.submitted} transformError={transformError} >
            <Part1 state={state} actions={actions} />
            <Part2 state={state} actions={actions} />
            <Part3 state={state} actions={actions} />
        </DataView>
    </div>
);

export default connect(
    (state) => ({ state: state.pages.uploader, appstate: state }),
    (dispatch) => ({
        actions: bindActionCreators(Actions, dispatch),
        go: (v) => dispatch(go(v)),
        clearTransform: () => dispatch(setSearchSubmit("")),
        transformError: (txt) => dispatch(setSearchState({ error: txt }))
    })
)(Render);