/*
  This is the window shown on dropdown for certain views that allows showing a quick description
  of the view's backend
*/

import React, {PropTypes} from 'react';

import TransformInput from '../../../components/TransformInput';

export default function generateDropdownTransformDisplay(description, transform) {
    return React.createClass({
        render: function() {
            let tf = transform;
            if (this.props.state.transform !== undefined) {
                tf = this.props.state.transform;
            }
            let desc = description;
            if (this.props.state.description !== undefined) {
                desc = this.props.state.description;
            }
            return (
                <div>
                    <p>{desc}</p>
                    <h4 style={{
                        paddingTop: "10px"
                    }}>Transform</h4>
                    <p>This is the transform used to generate this visualization:</p>
                    <TransformInput transform={tf} onChange= { (txt) => null }/>
                </div>
            );
        }
    });
}
