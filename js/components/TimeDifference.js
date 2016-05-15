// Displays an element that shows the time elapsed form the gven timestamp
import React, {Component, PropTypes} from 'react';

import prettydate from 'pretty-date';

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
        this.timer = setInterval(() => this.tick(), 1000);
        this.setState({
            elapsed: prettydate.format(new Date(this.props.timestamp))
        });
    }
    componentWillReceiveProps(nextProps) {
        this.setState({
            elapsed: prettydate.format(new Date(nextProps.timestamp))
        });
    }

    tick() {
        this.setState({
            elapsed: prettydate.format(new Date(this.props.timestamp))
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
