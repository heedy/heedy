// connectStorage performs the necessary work to connect the given params to actual user/device/stream
// values. Ie, <ConnectStorage user="test"><User /></ConnectStorage> will give the value of user test
// to the User component.
import React, {Component, PropTypes} from 'react';
import storage from './storage';

const NoQueryIfWithinMilliseconds = 1000;

export default function connectStorage(Component) {
    return React.createClass({
        propTypes: {
            user: PropTypes.string,
            device: PropTypes.string,
            stream: PropTypes.string,
            params: PropTypes.shape({user: PropTypes.string, device: PropTypes.string, stream: PropTypes.string})
        },
        getUser(props) {
            if (props === undefined) {
                props = this.props;
            }
            if (props.user !== undefined) 
                return props.user;
            if (props.params.user !== undefined) 
                return props.params.user;
            return "";
        },
        getDevice(props) {
            if (props === undefined) {
                props = this.props;
            }
            if (props.device !== undefined) 
                return props.device;
            if (props.params.device !== undefined) 
                return props.params.device;
            return "";
        },
        getStream(props) {
            if (props === undefined) {
                props = this.props;
            }
            if (props.stream !== undefined) 
                return props.stream;
            if (props.params.stream !== undefined) 
                return props.params.stream;
            return "";
        },
        getInitialState: function() {
            return {user: null, device: null, stream: null, error: null};
        },
        getData: function(nextProps) {
            let thisUser = this.getUser(nextProps);
            // Get the user/device/stream from cache - this allows the app to feel fast in
            // slow internet, and enables working in offline mode
            storage.get(thisUser).then((response) => {

                if (response != null) {
                    if (response.ref !== undefined) {
                        this.setState({error: response});
                    } else if (response.name !== undefined) {
                        this.setState({user: response});
                        // If the user was recently queried, don't query it again needlessly
                        if (response.timestamp > Date.now() - NoQueryIfWithinMilliseconds) {
                            return;
                        }
                    }
                }
                // The query will be caught by the callback
                storage.query(thisUser).catch((err) => console.log(err));
            });
            if (this.getDevice(nextProps) != "") {
                let thisDevice = thisUser + "/" + this.getDevice(nextProps);
                storage.get(thisDevice).then((response) => {
                    if (response != null) {
                        if (response.ref !== undefined) {
                            this.setState({error: response});
                        } else if (response.name !== undefined) {
                            this.setState({device: response});
                            // If the user was recently queried, don't query it again needlessly
                            if (response.timestamp > Date.now() - NoQueryIfWithinMilliseconds) {
                                return;
                            }
                        }
                    }
                    // The query will be caught by the callback
                    storage.query(thisDevice).catch((err) => console.log(err));
                });
                if (this.getStream(nextProps) != "") {
                    let thisStream = thisDevice + "/" + this.getStream(nextProps);
                    storage.get(thisStream).then((response) => {
                        if (response != null) {
                            if (response.ref !== undefined) {
                                this.setState({error: response});
                            } else if (response.name !== undefined) {
                                this.setState({stream: response});
                                // If the user was recently queried, don't query it again needlessly
                                if (response.timestamp > Date.now() - NoQueryIfWithinMilliseconds) {
                                    return;
                                }
                            }
                        }
                        // The query will be caught by the callback
                        storage.query(thisStream).catch((err) => console.log(err));
                    });
                }
            }
        },
        componentWillMount: function() {
            // https://stackoverflow.com/questions/1349404/generate-a-string-of-5-random-characters-in-javascript
            this.callbackID = Math.random().toString(36).substring(7);

            // Add the callback for storage
            storage.addCallback(this.callbackID, (path, obj) => {
                // If the current user/device/stream was updated, update the view
                if (path == this.getUser()) {
                    if (obj.ref !== undefined) {
                        this.setState({error: obj});
                    } else {
                        this.setState({user: obj});
                    }
                } else if (path == this.getDevice()) {
                    if (obj.ref !== undefined) {
                        this.setState({error: obj});
                    } else {
                        this.setState({device: obj});
                    }
                } else if (path == this.getStream()) {
                    if (obj.ref !== undefined) {
                        this.setState({error: obj});
                    } else {
                        this.setState({stream: obj});
                    }
                }
            });
            this.getData(this.props);
        },
        componentWillUnmount() {
            storage.remCallback(this.callbackID);
        },
        componentWillReceiveProps(nextProps) {
            if (this.getUser() != this.getUser(nextProps) || this.getDevice() != this.getDevice(nextProps) || this.getStream() != this.getStream(nextProps)) {
                this.setState({user: null, device: null, stream: null, error: null});

                this.getData(nextProps);
            }
        },
        render: function() {
            return (<Component {...this.props} user={this.state.user} device={this.state.device} stream={this.state.stream} error={this.state.error}/>);
        }
    });
}
