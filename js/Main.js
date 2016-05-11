import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

class Main extends Component {
    static propTypes = {};

    render() {
        console.log("Render main");
        return (
            <p>
                Main
            </p>
        );
    }
}

export default Main;
