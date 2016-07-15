/*
  Logout logs out of the app. It is a component, rather than direct navigation to /logout, because the frontend uses
  lots of locally cached data. This data needs to be cleared before we can log out, so that we don't leave behind any
  information.
*/

import React, {Component, PropTypes} from 'react';
import storage from './storage'

class Logout extends Component {
    constructor(props) {
        storage.clear().then(() => {
            console.log("Cleared local storage");
            // Navigate to the logout of the ConnectorDB server, which will remove cookies
            window.location = SiteURL + "/logout";
        }).catch((err) => {
            alert("Failed to clear local storage: " + err);
            window.location = SiteURL + "/logout";
        });
        super(props);
    }
    render() {
        return (
            <div style={{
                textAlign: "center",
                paddingTop: 200
            }}>
                <h1>
                    Logging Out ...
                </h1>
                <p>Clearing local data...</p>
            </div>
        );
    }
}

export default Logout;
