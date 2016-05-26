import React, {Component, PropTypes} from 'react';
import datatypes from './datatypes'

// TEMPORARY HACK: react-star-rating package is outdated on npm with an old react version.
// github has correct version, which needs to be compiled and stuff, so I just included the
// compiled files while waiting for an update to npm
//import StarRating from 'react-star-rating';
//import 'react-star-rating/dist/css/react-star-rating.min.css';
import StarRating from './react-star-rating/react-star-rating.min';
import './react-star-rating/react-star-rating.min.css';

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
        onChange: PropTypes.func,
        onSubmit: PropTypes.func
    }
    render() {
        let value = this.props.state.value;
        if (value === undefined || value == null)
            value = 0;

        // rating={value} messes up our ability to set the rating again in current version of react star rating. We therefore can't have it set :(
        return (<StarRating name={this.props.path} totalStars={10} size={30} onRatingClick={(a, val) => {
            console.log("Changing value:", val);
            this.props.onChange({value: val["rating"]});
            this.props.onSubmit(val["rating"], false);
        }}/>);
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
