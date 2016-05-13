import React, {Component, PropTypes} from 'react';
import storage from './storage'

class Logout extends Component {
    constructor(props) {
        storage.clear().then(() => {
            console.log("Cleared local storage");
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
