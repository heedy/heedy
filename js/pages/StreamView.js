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

import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import StreamCard from '../components/StreamCard';
import DataInput from '../components/DataInput';
import DataQuery from '../components/DataQuery';
import DataViewCard from '../components/DataViewCard';
import SearchCard from '../components/SearchCard';

import {getViews} from '../datatypes/datatypes';

import {setSearchSubmit, setSearchState} from '../actions';

class StreamView extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        device: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        stream: PropTypes.object.isRequired,
        state: PropTypes.shape({expanded: PropTypes.bool.isRequired}).isRequired,
        thisUser: PropTypes.object.isRequired,
        thisDevice: PropTypes.object.isRequired,
        pipescript: PropTypes.object
    }

    // Sets up the transform (if any) so that we're good to go with pipescript
    initTransform(t) {
        if (t !== undefined && t !== "") {
            if (this.props.pipescript != null) {
                try {
                    this.transform = this.props.pipescript.Script(t);
                    this.props.transformError("");
                    return;
                } catch (e) {
                    this.props.transformError(e.toString());
                }

            }
        }
        if (t === "") {
            this.props.transformError("");
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

    // Generate the component's initial state
    componentWillMount() {
        // We use the value of the submitted search to define our transform
        this.initTransform(this.props.state.search.submitted);

        this.data = this.dataTransform(this.props.state.data);
    }

    // Each time either the data or the transform changes, reload
    componentWillReceiveProps(p) {
        // We only perform the dataset transform operation if the dataset
        // was modified
        if (p.state.data !== this.props.state.data || this.props.state.search.submitted !== p.state.search.submitted) {
            if (this.props.state.search.submitted !== p.state.search.submitted) {
                this.initTransform(p.state.search.submitted);
            }
            this.data = this.dataTransform(p.state.data);
        }
    }

    render() {
        let state = this.props.state;
        let user = this.props.user;
        let device = this.props.device;
        let stream = this.props.stream;
        let streamschema = JSON.parse(stream.schema);

        // Finally, we check what views to show
        let views = getViews({
            data: this.data,
            user: user,
            device: device,
            stream: stream,
            schema: streamschema,
            pipescript: this.props.pipescript,
            thisUser: this.props.thisUser,
            thisDevice: this.props.thisDevice
        });

        return (
            <div>
                {this.transform != null
                    ? (<SearchCard title={state.search.submitted} subtitle={"Transform applied to data"} onClose={() => this.props.clearTransform()}/>)
                    : null}
                <StreamCard user={user} device={device} stream={stream} state={state}/>

                <div style={{
                    marginLeft: "-15px",
                    marginRight: "-15px"
                }}>
                    {stream.downlink || this.props.thisUser.name == user.name && this.props.thisDevice.name == device.name
                        ? (<DataInput user={user} device={device} stream={stream} schema={streamschema} thisUser={this.props.thisUser} thisDevice={this.props.thisDevice}/>)
                        : null}
                    <DataQuery state={state} user={user} device={device} stream={stream}/> {views.map((view) => {
                        return (<DataViewCard key={view.key} view={view} data={this.data} user={user} device={device} stream={stream} schema={streamschema} state={state} pipescript={this.props.pipescript} thisUser={this.props.thisUser} thisDevice={this.props.thisDevice}/>);
                    })}
                </div>
            </div>
        );
    }
}

export default connect((state) => ({thisUser: state.site.thisUser, thisDevice: state.site.thisDevice, pipescript: state.site.pipescript}), (dispatch) => ({
    clearTransform: () => dispatch(setSearchSubmit("")),
    transformError: (txt) => dispatch(setSearchState({error: txt}))
}))(StreamView);
