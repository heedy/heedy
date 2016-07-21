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
import {showMessage} from '../actions';

import ExpandableCard from './ExpandableCard';

class DataViewCard extends Component {
    static propTypes = {
        view: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        thisUser: PropTypes.object.isRequired,
        thisDevice: PropTypes.object.isRequired,
        msg: PropTypes.func.isRequired,
        setState: PropTypes.func.isRequired
    }

    render() {
        let view = this.props.view;
        let state = this.props.state;
        let setState = this.props.setState;
        let curstate = (state.views[view.key] !== undefined
            ? state.views[view.key]
            : view.initialState);

        let context = {
            user: this.props.user,
            device: this.props.device,
            stream: this.props.stream,
            state: curstate,
            thisUser: this.props.thisUser,
            thisDevice: this.props.thisDevice,
            msg: this.props.msg,
            data: state.data,
            setState: (v) => {
                let newViews = Object.assign({}, state.views);
                newViews[view.key] = Object.assign({}, curstate, v);
                setState({views: newViews});
            }
        };
        let dropdown = null;
        if (view.dropdown !== undefined) {
            dropdown = (<view.dropdown {...context}/>);
        }

        return (
            <ExpandableCard width={view.width} state={curstate} setState={context.setState} dropdown={dropdown} title={view.title} subtitle={view.subtitle} style={view.style}>
                <view.component {...context}/>
            </ExpandableCard>
        );
    }
}
export default connect(undefined, (dispatch, props) => {
    let path = props.user.name + "/" + props.device.name + "/" + props.stream.name;
    return {
        setState: (s) => dispatch({type: "STREAM_VIEW_SET", name: path, value: s}),
        msg: (t) => dispatch(showMessage(t))
    };
})(DataViewCard);
