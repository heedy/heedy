/**
  The StreamView represents the main page shown when viewing a stream. It consists of 2 parts
  parts:
   - header card: the same as device/user cards - it is expandable to show stream details as well
   as show icons to edit the stream
   - The main bootstrap grid, which contains all visualization/querying of stream data.

   The rest of this comment pertain to the bootstrap grid, which is this page's main event.

   The grid is made up of 3 distinct parts:
  - If the stream is a downlink stream, or the stream is the current device, show the data input control.
      This corresponds to the default permissions structure where the owner alone can write streams unless
      they are explicitly marked as downlink.
  - The main data query control - allows you to query data from your stream however you want. The queried data
      will be what is shown in the next part (the analysis cards)
  - Analysis cards. These cards are given the data which is queried in the main query control, and plot/show
      visualizations. Each analysis card specifies if it is to be full-width, half-width, or expandable and gives an optional
      control area that drops down. Expandable cards can switch between full and half-width based on user input.

  While there is no queried data, the analysis cards are replaced with a loading screen.
**/

import React, { Component, PropTypes } from 'react';
import { connect } from 'react-redux';

import StreamCard from '../components/StreamCard';
import DataInput from '../components/DataInput';
import DataQuery from '../components/DataQuery';
import DataView from '../components/DataView';
import SearchCard from '../components/SearchCard';

import { setSearchSubmit, setSearchState } from '../actions';

const StreamView = ({ user, device, stream, state, thisUser, thisDevice, clearTransform, transformError }) => (
    <div>
        {state.search.submitted != null && state.search.submitted != ""
            ? (<SearchCard title={state.search.submitted} subtitle={"Transform applied to data"} onClose={clearTransform} />)
            : null}
        <StreamCard user={user} device={device} stream={stream} state={state} />

        <DataView data={state.data} transform={state.search.submitted} transformError={transformError} datatype={stream.datatype} schema={JSON.parse(stream.schema)} >
            {stream.downlink || thisUser.name == user.name && thisDevice.name == device.name
                ? (<DataInput user={user} device={device} stream={stream} schema={JSON.parse(stream.schema)} thisUser={thisUser} thisDevice={thisDevice} />)
                : null}
            <DataQuery state={state} user={user} device={device} stream={stream} />
        </DataView>
    </div>
);


export default connect((state) => ({ thisUser: state.site.thisUser, thisDevice: state.site.thisDevice }), (dispatch) => ({
    clearTransform: () => dispatch(setSearchSubmit("")),
    transformError: (txt) => dispatch(setSearchState({ error: txt }))
}))(StreamView);
