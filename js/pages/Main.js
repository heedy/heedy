/*
  Main is the index page shown after initially logging in to ConnectorDB
*/

import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {go} from '../actions';

import MainToolbar from '../components/MainToolbar';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import Welcome from '../components/Welcome';
import DataInput from '../components/DataInput';

class Main extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        device: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        streamarray: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,

        onStreamClick: PropTypes.func.isRequired
    }

    render() {
        let state = this.props.state;
        let user = this.props.user;
        let device = this.props.device;
        let streams = this.props.streamarray;
        return (
            <div style={{
                textAlign: "left"
            }}>
                <MainToolbar user={user} device={device} state={state}/> {streams == null || Object.keys(streams).length == 0
                    ? (<Welcome/>)
                    : (
                        <div style={{
                            marginLeft: "-15px",
                            marginRight: "-15px"
                        }}>{Object.keys(streams).map((skey) => {
                                let s = streams[skey];
                                let path = user.name + "/" + device.name + "/" + s.name;
                                return (

                                    <DataInput key={s.name} size={6} title={s.nickname == ""
                                        ? s.name
                                        : s.nickname} subtitle={path} user={user} device={device} stream={s}>
                                        <div style={{
                                            float: "right",
                                            marginTop: "-5px",
                                            marginLeft: "-100px"
                                        }}>
                                            <IconButton onTouchTap={() => this.props.onStreamClick(path)} tooltip="view stream">
                                                <FontIcon className="material-icons" color="rgba(0,0,0,0.5)">
                                                    list
                                                </FontIcon>
                                            </IconButton>
                                        </div>
                                    </DataInput>
                                );
                            })}</div>
                    )}

            </div>
        );
    }
}
export default connect(undefined, (dispatch, props) => ({
    onStreamClick: (s) => dispatch(go(s))
}))(Main);
