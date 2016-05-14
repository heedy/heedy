import React, {Component, PropTypes} from 'react';
import CircularProgress from 'material-ui/CircularProgress';

export default function Loading() {
    return (
        <div style={{
            textAlign: "center",
            paddingTop: 200
        }}>
            <CircularProgress/>
        </div>
    );
}
