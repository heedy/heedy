// connectStorage performs the necessary work to connect the given params to actual user/device/stream
// values. Ie, <ConnectStorage user="test"><User /></ConnectStorage> will give the value of user test
// to the User component.

// TODO: I cry when I see code like this. What makes it all the more horrible is that *I* am
//  the person who wrote it... This really needs to be refactored... - dkumor

import React, {Component, PropTypes} from 'react';
import storage from './storage';

const NoQueryIfWithinMilliseconds = 1000;

export default function connectStorage(Component, lsdev, lsstream) {
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
            return {
                user: null,
                device: null,
                stream: null,
                error: null,
                devarray: null,
                streamarray: null
            };
        },
        getData: function(nextProps) {
            var thisUser = this.getUser(nextProps);
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
                var thisDevice = thisUser + "/" + this.getDevice(nextProps);
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
                    var thisStream = thisDevice + "/" + this.getStream(nextProps);
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
            //Whether or not to add lists of children
            if (lsdev) {
                storage.ls(thisUser).then((response) => {
                    if (response.ref !== undefined) {
                        this.setState({error: response});
                    } else {
                        this.setState({devarray: response});

                    }
                    // The query will be caught by the callback
                    storage.query_ls(thisUser).catch((err) => console.log(err));
                })
            }
            if (lsstream) {
                storage.ls(thisDevice).then((response) => {
                    if (response.ref !== undefined) {
                        this.setState({error: response});
                    } else {
                        this.setState({streamarray: response});

                    }
                    // The query will be caught by the callback
                    storage.query_ls(thisDevice).catch((err) => console.log(err));
                })
            }
        },
        componentWillMount: function() {
            // https://stackoverflow.com/questions/1349404/generate-a-string-of-5-random-characters-in-javascript
            this.callbackID = Math.random().toString(36).substring(7);

            // Add the callback for storage
            storage.addCallback(this.callbackID, (path, obj) => {
                // If the current user/device/stream was updated, update the view
                let thisUser = this.getUser();
                let thisDevice = thisUser + "/" + this.getDevice();
                let thisStream = thisDevice + "/" + this.getStream();
                if (path == thisUser) {
                    if (obj.ref !== undefined) {
                        this.setState({error: obj});
                    } else {
                        this.setState({user: obj});
                    }
                } else if (path == thisDevice) {
                    if (obj.ref !== undefined) {
                        this.setState({error: obj});
                    } else {
                        this.setState({device: obj});
                    }
                } else if (path == thisStream) {
                    if (obj.ref !== undefined) {
                        this.setState({error: obj});
                    } else {
                        this.setState({stream: obj});
                    }
                } else if ((lsdev || lsstream) && obj.ref === undefined && path.startsWith(thisUser + "/")) {
                    // We might want to update our arrays
                    let p = path.split("/");
                    switch (p.length) {
                        case 2:
                            if (lsdev) {
                                let ndevarray = Object.assign({}, this.state.devarray);
                                ndevarray[path] = obj;
                                this.setState({devarray: ndevarray});
                            }
                            break;
                        case 3:
                            if (p[1] == this.getDevice() && lsstream) {

                                let nsarray = Object.assign({}, this.state.streamarray);
                                nsarray[path] = obj;
                                this.setState({streamarray: nsarray});
                            }
                            break;
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
                this.setState({
                    user: null,
                    device: null,
                    stream: null,
                    devarray: null,
                    streamarray: null,
                    error: null
                });

                this.getData(nextProps);
            }
        },
        render: function() {
            return (<Component {...this.props} user={this.state.user} device={this.state.device} stream={this.state.stream} error={this.state.error} devarray={this.state.devarray} streamarray={this.state.streamarray}/>);
        }
    });
}
