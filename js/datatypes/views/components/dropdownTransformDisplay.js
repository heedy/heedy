/*
  This is the window shown on dropdown for certain views that allows showing a quick description
  of the view's backend
*/

import React, {PropTypes} from 'react';

import TransformInput from '../../../components/TransformInput';

export default function generateDropdownTransformDisplay(description, transform) {
    return React.createClass({
        render: function() {
            return (
                <div>
                    <p>{description}</p>
                    <h4 style={{
                        paddingTop: "10px"
                    }}>Transform</h4>
                    <p>This is the transform used to generate your visualization:</p>
                    <TransformInput transform={transform} onChange= { (txt) => null }/>
                </div>
            );
        }
    });
}
