# inputs

The following is a stub for creating a new input:

```javascript
/**
A comment mentioning what your input does.
**/

import React, {Component, PropTypes} from 'react';

// depending on where your code is located, you might need to modify the import location
import {addInput} from '../datatypes';


class MyInput extends Component {
    static propTypes = {
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        path: PropTypes.string.isRequired, // myuser/mydevice/mystream
        schema: PropTypes.object.isRequired, // The schema javascript object
        state: PropTypes.object.isRequired, // The state - an object.
        onSubmit: PropTypes.func.isRequired, // Send in the data portion of your datapoint here
        setState: PropTypes.func.isRequired,  // Allows you to set the state
        showMessage: PropTypes.func.isRequired // Allows you to show error messages
    }

    render() {
      return (
        <div>
          <p>Hello World!</p>
        </div>
      )
    }
}

// add the input to the input registry. here, "hello.world" is the datatype
addInput("hello.world", {
    width: "expandable-half", // We want the input to be half-width by default, but expandable
    component: MyInput
});

```
