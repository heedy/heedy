import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {setTitle} from './util';

class Main extends Component {
    static propTypes = {};
    componentDidMount() {
        setTitle("");
    }
    componentWillReceiveProps(newProps) {
        setTitle("");
    }
    render() {
        return (
            <p>
                Main
            </p>
        );
    }
}

export default Main;
