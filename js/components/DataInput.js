import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {Card, CardText, CardHeader} from 'material-ui/Card';

import Form from "react-jsonschema-form";

import {dataInput, showMessage} from '../actions';
import {go} from '../actions';

const log = (type) => console.log.bind(console, type);

// Unfortunately the schema form generator is... kinda BS in that it has undefined for values. To fix that,
// we modify the schema
function generateSchema(s) {
    let uiSchema = {};
    switch (s.type) {
        case "object":
            let k = Object.keys(s.properties);
            for (let i in k) {
                let key = k[i];
                let ret = generateSchema(s.properties[key]);
                uiSchema[key] = ret.ui;
                s.properties[key] = ret.s;
            }
            break;
        case "string":
            if (s.default === undefined) {
                s["default"] = "";
            }

            break;
        case "boolean":
            if (s.default === undefined) {
                s["default"] = false;
            }
            uiSchema["ui:widget"] = "radio";
            break;
        case "number":
            if (s.default === undefined) {
                s["default"] = 0;
            }
            break;

    }
    return {ui: uiSchema, s: s};
}

const noSchema = {
    type: "object",
    properties: {
        input: {
            title: "Stream Data JSON",
            type: "string"
        }
    }
};

class DataInput extends Component {
    static propTypes = {
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        onSubmit: PropTypes.func.isRequired,
        showMessage: PropTypes.func.isRequired,
        title: PropTypes.string,
        subtitle: PropTypes.string
    }

    static defaultProps = {
        title: "Input Data to Downlink",
        subtitle: ""
    }

    touch() {
        if (this.props.touch !== undefined) {
            this.props.touch();
        }
    }
    render() {
        let user = this.props.user;
        let device = this.props.device;
        let stream = this.props.stream;

        let schema = JSON.parse(stream.schema);
        let curschema = Object.assign({}, schema);
        let inside = false;

        if (schema.type === undefined) {
            curschema = noSchema;
        } else if (schema.type != "object") {
            if (curschema.title === undefined) {
                curschema.title = "Input Data:"
            }
            curschema = {
                type: "object",
                properties: {
                    input: curschema
                }

            };
        }
        let s = generateSchema(curschema);

        return (
            <div className="col-lg-6">
                <Card style={{
                    marginTop: "20px",
                    textAlign: "left"
                }}>
                    <CardHeader title={this.props.title} subtitle={this.props.subtitle}>{this.props.children}</CardHeader>
                    <CardText style={{
                        textAlign: "center"
                    }}>
                        <Form schema={s.s} uiSchema={s.ui} onSubmit={(data) => {
                            if (schema.type === undefined) {
                                try {
                                    var parsedData = JSON.parse(data.formData.input);
                                } catch (e) {
                                    this.props.showMessage(e.toString());
                                    return;
                                }
                                this.props.onSubmit(parsedData);
                            } else if (schema.type != "object") {
                                this.props.onSubmit(data.formData.input);
                                return;
                            }
                            this.props.onSubmit(data.formData);
                        }} onError={log("errors")}/>
                    </CardText>
                </Card>
            </div>
        );
    }
}

export default connect((state) => ({}), (dispatch, props) => ({
    onSubmit: (val) => dispatch(dataInput(props.user, props.device, props.stream, val)),
    showMessage: (val) => dispatch(showMessage(val))
}))(DataInput);
