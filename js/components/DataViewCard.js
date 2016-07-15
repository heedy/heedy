/**
  DataViewCard is the card which holds a "data view" - it is given stream details as well as
  the array of datapoints that was queried, and a "view", which is defined in the datatypes folder.
  You can find examples of this card when looking at stream data - all plots and data tables are "views" rendered
  within DataViewCards.

  The DataViewCard sets up the view and displays it within its card. It manages the display size
  of the card, as well as managing the extra dropdown options (if given). This greatly simplifies
  the repeated code used in each view.
**/

import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import FlatButton from 'material-ui/FlatButton';

import TextField from 'material-ui/TextField';

import DataTable from './DataTable';

import DateTime from 'react-datetime';
import 'react-datetime/css/react-datetime.css';

import {query, showMessage} from '../actions';

class DataViewCard extends Component {
    static propTypes = {
        view: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        msg: PropTypes.func.isRequired,
        setState: PropTypes.func.isRequired
    }

    render() {
        let view = this.props.view;
        let state = this.props.state;
        let setState = this.props.setState;

        // Whether to show the dropdown button and set up the dropdown component
        // for this view
        let Dropdown = view.dropdown;
        let hasDropdown = (Dropdown !== undefined && Dropdown !== null);

        // The main view component
        let View = view.view;

        let expandable = view.width.startsWith("expandable");
        let width = view.width;
        if (expandable) {
            switch (width) {
                case "expandable":
                    width = "half";
                    break;
                case "expandable-half":
                    width = "half";
                    break;
                case "expandable-full"
                    width = "full";
                    break;
            }
        }
        // The state's width overrides built-in values
        if (state.width !== undefined) {
            width = state.width;
        }

        return (
            <div className={width === "full"
                ? "col-lg-12"
                : "col-lg-6"}>
                <Card style={{
                    marginTop: "20px",
                    textAlign: "left"
                }} onExpandChange={(val) => setState({
                    ...state,
                    expanded: val
                })} expanded={state.expanded}>
                    <CardHeader title={state.title === undefined
                        ? view.title
                        : state.title} showExpandableButton={hasDropdown}>
                        <div style={{
                            float: "right",
                            marginRight: 25,
                            marginTop: "-15px",
                            marginLeft: "-300px"
                        }}>
                            {state.fullwidth
                                ? (
                                    <IconButton onTouchTap= { (val) => setState({ ...state, width: "half" }) }>
                                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                            call_received
                                        </FontIcon>
                                    </IconButton>

                                )
                                : (

                                    <IconButton onTouchTap= { (val) => setState({ ...state, width: "full"}) }>
                                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                            call_made
                                        </FontIcon >
                                    </IconButton>
                                )}
                        </div>
                    </CardHeader>
                    {hasDropdown
                        ? (
                            <CardText expandable={true} style={{
                                backgroundColor: "rgba(0,179,74,0.05)",
                                paddingBottom: "30px"
                            }}>
                                <Dropdown ...this.props/>
                            </CardText>
                        )
                        : null}

                    <CardText>
                        <View ...this.props/>
                    </CardText>
                </Card>
            </div>
        );
    }
}
export default connect(undefined, (dispatch, props) => {
    let path = props.user.name + "/" + props.device.name + " / " + props.stream.name;
    return {
        query: (q) => dispatch(query(props.user, props.device, props.stream, q)),
        setState: (s) => dispatch({type: "STREAM_VIEW_SET", name: path, value: s}),
        msg: (t) => dispatch(showMessage(t))
    };
})(DataViewCard);
