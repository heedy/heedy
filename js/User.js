import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import storage from './storage';

import Error from './Error';
import Loading from './Loading';

class User extends Component {
    static propTypes = {
        params: PropTypes.shape({user: PropTypes.string.isRequired}).isRequired
    };

    constructor(props) {
        super(props);
        this.state = {
            user: null,
            error: null
        };
        this.isquerying = false;
    }

    // Upon mounting, create a callback which will update the component state when the user is updated
    componentDidMount() {
        // https://stackoverflow.com/questions/1349404/generate-a-string-of-5-random-characters-in-javascript
        this.callbackID = Math.random().toString(36).substring(7);

        storage.addCallback(this.callbackID, (path, obj) => {
            // If the current user was updated, update the view
            if (path == this.props.params.user) {
                if (obj.ref !== undefined) {
                    this.setState({error: obj});
                } else {
                    this.setState({user: obj});
                }
            }
        });

        // componentWillReceiveProps is not called on mount for some reason... so let's do this
        this.componentWillReceiveProps(this.props);
    }
    componentWillUnmount() {
        storage.remCallback(this.callbackID);
    }

    // This is called whenever the user changes
    componentWillReceiveProps(nextProps) {
        if (this.isquerying) {
            return;
        }
        this.isquerying = true;
        // Set the user to null, since the old user should not longer be displayed
        this.setState({user: null, error: null});

        // Get the user from cache - this allows the app to feel fast on slow internet,
        // and enables it to work in offline mode
        storage.get(nextProps.params.user).then((response) => {
            this.isquerying = false;
            // Set the user to the cached value
            if (response != null) {
                if (response.ref !== undefined) {
                    this.setState({error: response});
                } else if (response.name !== undefined) {
                    this.setState({user: response});

                    // If the user was recently queried, don't query it again needlessly
                    if (response.timestamp > Date.now() - 1000 * 10) {
                        return;
                    }
                }
            }

            storage.query(nextProps.params.user).catch((err) => console.log(err));
        }).catch((err) => {
            this.isquerying = false;
            console.log(err);
        });
    }

    render() {
        if (this.state.error != null) {
            // There was an error
            return (<Error err={this.state.error}/>);
        }
        if (this.state.user == null) {
            // The user is currently being queried
            return (<Loading/>);
        }

        return (
            <div>
                <h1>User: {this.state.user.name}</h1>
                <h3>{this.state.user.timestamp}</h3>
            </div>
        );
    }
}

export default User;
