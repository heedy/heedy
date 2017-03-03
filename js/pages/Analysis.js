import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';


import * as Actions from '../actions/analysis';

const Render = ({ state, actions }) => (
    <div style={{
        textAlign: "left"
    }}>

    </div>
);

export default connect(
    (state) => ({ state: state.pages.analysis }),
    (dispatch) => ({ actions: bindActionCreators(Actions, dispatch) })
)(Render);