import React, {Component, PropTypes} from 'react';
import datatypes from './datatypes'

export const ratingSchema = {
    type: "integer",
    minimum: 0,
    maximum: 10
}

class RatingCreate extends Component {
    static propTypes = {
        state: PropTypes.object,
        callbacks: PropTypes.object
    }
    render() {
        return (null);
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
    input: {
        component: DataInput,
        size: 1, // One of 1 or 2 meaning normal or double size of the data input card
    },
    create: {
        required: null,
        optional: null,
        description: "A rating allows you to manually rate things such as your mood or productivity out of 10 stars. Ratings are your way of telling ConnectorDB how you think your life is going.",
        default: {
            schema: JSON.stringify(ratingSchema),
            datatype: "rating.stars",
            ephemeral: false,
            downlink: false
        }
    },
    name: "rating"
};
