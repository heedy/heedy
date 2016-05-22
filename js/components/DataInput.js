import React, {Component, PropTypes} from 'react';
import {Card, CardText, CardHeader} from 'material-ui/Card';

import Form from "react-jsonschema-form";
const log = (type) => console.log.bind(console, type);

const schema = {
    type: "object",
    properties: {
        input: {
            title: "Stream Data JSON",
            type: "string"
        }
    }
};

const formData = {};

class DataInput extends Component {
    render() {
        return (
            <Card style={{
                marginTop: "20px",
                textAlign: "left"
            }}>
                <CardHeader title={"Input Data To Downlink"}/>
                <CardText style={{
                    textAlign: "center"
                }}>
                    <Form schema={schema} formData={formData} onChange={log("changed")} onSubmit={log("submitted")} onError={log("errors")}/>
                </CardText>
            </Card>
        );
    }
}
export default DataInput;
