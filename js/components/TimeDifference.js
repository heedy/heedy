// Displays an element that shows the time elapsed form the gven timestamp
import React, {Component, PropTypes} from 'react';

import moment from 'moment';

export default class TimeDifference extends Component {
    static propTypes = {
        timestamp: PropTypes.number.isRequired
    }
    constructor(props) {
        super(props);
        this.state = {
            elapsed: ""
        }
    }
    componentWillMount() {
        this.timer = setInterval(() => this.tick(), 60 * 1000);
        this.setState({
            elapsed: moment(new Date(this.props.timestamp)).fromNow()
        });
    }
    componentWillReceiveProps(nextProps) {
        this.setState({
            elapsed: moment(new Date(nextProps.timestamp)).fromNow()
        });
    }

    tick() {
        this.setState({
            elapsed: moment(new Date(this.props.timestamp)).fromNow()
        });
    }
    componentWillUnmount() {
        clearInterval(this.timer);
    }
    render() {
        return ( < div > {
            this.state.elapsed
        } < /div>);
    }
}
