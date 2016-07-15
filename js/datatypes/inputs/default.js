/**
The default input - it is returned if no custom inputs are available for the given
datatype.

This input uses the 'react-jsonschema-form' library to create an input form which fits
the stream's schema. If the stream has no schema, it has a textbox in which the user may
type in arbitrary JSON.
**/

import React, {Component, PropTypes} from 'react';

import Form from "react-jsonschema-form";

import addInput from '../datatypes';

// The schema to use for generating the form when no schema is specified
// in the stream
const noSchema = {
    type: "object",
    properties: {
        input: {
            title: "Stream Data JSON",
            type: "string"
        }
    }
};

// Unfortunately the schema form generator is... kinda BS in that it has
// undefined defaults for values. The form generator also doesn't do a good job
// handling non-object schemas. This function does two things:
// 1) It modifies the schema given to include default values and be ready for input
// 2) It generates a uischema, which allows us to set specific view types for
//  certain schemas. Currently it is used to generate booleans as radio buttons.
function prepareSchema(s) {
    let uiSchema = {};
    let schema = Object.assign({}, s); // We'll be modifying the object, so copy it
    if (s.type === undefined) {
        schema = noSchema;
    } else {
        // The schema is valid - set up the default values and uiSchema
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

        // The form generator doesn't handle non-object schemas well, so if the
        // root type is not object, we wrap the schema in an object
        if (schema.type != "object") {
            if (schema.title === undefined) {
                curschema.title = "Input Data:"
            }
            schema = {
                type: "object",
                properties: {
                    input: schema
                }
            };
        }
    }

    return {ui: uiSchema, schema: schema};
}

class DefaultInput extends Component {
    static propTypes = {
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        stream: PropTypes.object.isRequired,
        schema: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        onSubmit: PropTypes.func.isRequired,
        setState: PropTypes.func.isRequired,
        showMessage: PropTypes.func.isRequired
    }

    // submit is run when the user clicks submit. It manages the different cases of data
    // that are managed by the input - no schema, wrapped schema, and normal schema :)
    submit(data) {
        let schema = this.props.schema;
        if (schema.type === undefined) {
            // If the stream has no schema, make sure we can parse the data as JSON
            // before inserting it.
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
    }

    render() {
        let user = this.props.user;
        let device = this.props.device;
        let stream = this.props.stream;
        let path = user.name + "/" + device.name + "/" + stream.name;

        let preparedSchema = prepareSchema(this.props.schema);
        let state = this.props.state;
        return (<Form schema={preparedSchema.schema} uiSchema={preparedSchema.ui} formData={state} onChange={this.props.setState} onSubmit={(data) => this.submit(data)} onError={log("error in the input form")}/>);

    }
}

// add the input to the input registry. The empty string makes it default
addInput("", {
    width: "expandable-half",
    component: DefaultInput
});
