import React, {Component, PropTypes} from 'react';
import datatypes from './datatypes'

class RatingCreate extends Component {
    static propTypes = {
        state: PropTypes.object,
        callbacks: PropTypes.object
    }
    render() {
        return (
            <p>Create rating!</p>
        );
    }
}

class DataInput extends Component {
    static propTypes = {
        state: PropTypes.object,
        onStateChange: PropTypes.func,
        onSubmit: PropTypes.func
    }
    render() {
        return (
            <p>Rating!</p>
        );
    }
}

// register the datatype
datatypes["rating.stars"] = {
    input: DataInput,
    create: RatingCreate
};

export default RatingCreate;
