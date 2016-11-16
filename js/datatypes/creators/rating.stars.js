import {addCreator} from '../datatypes';

export const ratingSchema = {
    type: "integer",
    minimum: 0,
    maximum: 10
};

addCreator("rating.stars", {
    name: "rating",
    required: null,
    optional: null,
    description: "A rating allows you to manually rate things such as your mood or productivity out of 10 stars. Ratings are your way of telling ConnectorDB how you think your life is going.",
    default: {
        schema: JSON.stringify(ratingSchema),
        datatype: "rating.stars",
        icon: "material:star",
        ephemeral: false,
        downlink: false
    }
});
