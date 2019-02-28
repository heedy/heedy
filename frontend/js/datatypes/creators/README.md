# creators

Creators are custom prompts to show when creating a specific datatype.

A good example is the star rating:

```javascript
/*
Comment describing your creator
*/
import {addCreator} from '../datatypes';

export const ratingSchema = {
    type: "integer",
    minimum: 0,
    maximum: 10
};

// addCreator is given a datatype, and you submit the following object:
addCreator("rating.stars", {
    // A react component to show in the required section
    required: null,
    // A react component to show in the optional section
    optional: null,
    // A description for the datatype, shown at the head
    description: "A rating allows you to manually rate things such as your mood or productivity out of 10 stars. Ratings are your way of telling ConnectorDB how you think your life is going.",

    // The default stream values
    default: {
        schema: JSON.stringify(ratingSchema),
        datatype: "rating.stars",
        ephemeral: false,
        downlink: false
    }
});

```
