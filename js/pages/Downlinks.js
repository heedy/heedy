import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { Card, CardText, CardHeader } from 'material-ui/Card';

import * as Actions from '../actions/downlinks';

const Welcome = () => (
    <Card style={{
        marginTop: "20px"
    }}>
        <CardHeader title={"Downlinks"} subtitle={"Control your devices through ConnectorDB"} />
        <CardText>
            <p>It looks like you don't have any downlinks set up yet. If you sync devices such as your lights or your thermostat to ConnectorDB, you will be able to control them here.</p>
            <p>Downlink streams allow external input, which is immediately sent to the relevant device. Once acknowledged by the device, the input is added to that stream's data.</p>

        </CardText>
    </Card>
);

const Render = ({ state, actions }) => (
    <div style={{
        textAlign: "left"
    }}>
        <Welcome />
    </div>
);

export default connect(
    (state) => ({ state: state }),
    (dispatch) => ({ actions: bindActionCreators(Actions, dispatch) })
)(Render);