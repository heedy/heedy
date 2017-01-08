import React, { Component, PropTypes } from 'react';
import QRCode from 'qrcode.react';

import { go } from '../actions';
import { app } from '../util';

const InitialState = {
    // roles represents the possible permissions allowed by ConnectorDB.
    // Note that the values given here correspond to the default ConnectorDB settings.
    // One can change ConnectorDB to have a different permissions structure, which would
    // make these values inconsistent with ConnectorDB.... so don't do that
    roles: {
        user: {
            user: {
                description: "can read/edit own devices and read public users/devices"
            },
            admin: {
                description: "has administrative access to the database"
            }
        },
        device: {
            user: {
                description: "has all permissions that the owning user has (Create/Read/Update/Delete)"
            },
            writer: {
                description: "can read and update basic properties of streams and devices"
            },
            reader: {
                description: "can read properties of streams and devices, and read their data"
            },
            none: {
                description: "the device is isolated - it only has access to itself and its own streams"
            }
        }
    },

    defaultschemas: [
        {
            description: "no schema (can insert anything)",
            name: "none",
            schema: {}
        }, {
            description: "number",
            name: "number",
            schema: {
                type: "number"
            }
        }, {
            description: "boolean",
            name: "boolean",
            schema: {
                type: "boolean"
            }
        }, {
            description: "string",
            name: "string",
            schema: {
                type: "string"
            }
        }
    ],

    // navigation is displayed in the app's main nmenu
    navigation: [
        {
            title: "Insert",
            subtitle: "Manually insert data",
            icon: "star",
            page: ""
        }, {
            title: "Devices",
            subtitle: "View your profile",
            icon: "devices",
            page: "{self}"
        }
    ],

    dropdownMenu: [
        {
            title: "Server Info",
            icon: "info_outline",
            action: (dispatch) => {
                dispatch({
                    type: "SHOW_DIALOG",
                    value: {
                        title: "Server Info",
                        open: true,
                        contents: (
                            <div>
                                <div className="col-sm-8">
                                    <h6>Server:</h6>
                                    <h3>{SiteURL}</h3>
                                    <br />
                                    <h6>Version:</h6>
                                    <h3>{ConnectorDBVersion}</h3>
                                </div>
                                <div className="col-sm-4" style={{ textAlign: "right" }}>
                                    <QRCode value={SiteURL} />
                                </div>
                            </div>
                        )
                    }
                });
            }
        },
        {
            title: "Documentation",
            icon: "help",
            action: (dispatch) => {
                window.location.href = "https://connectordb.io/docs/";
            }
        }, {
            title: "Submit Issue",
            icon: "bug_report",
            action: (dispatch) => {
                window.location.href = "https://github.com/connectordb/connectordb/issues";
            }
        }, {
            title: "Sign Out",
            icon: "power_settings_new",
            action: (dispatch) => {
                dispatch(go("logout"));
            }
        }
    ],

    // The currently logged in user and device. This is set up immediately on app start.
    // even before the app is rendered. Note that these are NOT updated along with
    // the app storage - this is the initial user and device
    thisUser: null,
    thisDevice: null,

    // The URL of the website, also available as global variable "SiteURL". This is set up
    // from the context on app load
    siteURL: "",

    // The status message to show in the snack bar
    status: "",
    statusvisible: false,

    // Show a modal dialog
    dialog: {
        title: "",
        contents: null,
        open: false
    },

    // Whether pipescript is loaded or not - this contains the pipescript library
    pipescript: null

};

export default function siteReducer(state = InitialState, action) {
    switch (action.type) {
        case 'LOAD_CONTEXT':
            var out = Object.assign({}, state, {
                siteURL: action.value.SiteURL,
                thisUser: action.value.ThisUser,
                thisDevice: action.value.ThisDevice
            });

            // now set up the navigation correctly (replace {self} with user name)
            for (var i = 0; i < out.navigation.length; i++) {
                out.navigation[i].title = out.navigation[i].title.replace("{self}", out.thisUser.name);
                out.navigation[i].subtitle = out.navigation[i].subtitle.replace("{self}", out.thisUser.name);
                out.navigation[i].page = out.navigation[i].page.replace("{self}", out.thisUser.name);
            }
            return out;
        case 'PIPESCRIPT':
            return {
                ...state,
                pipescript: action.value
            };
        case 'STATUS_HIDE':
            return {
                ...state,
                statusvisible: false
            };
        case 'SHOW_STATUS':
            return {
                ...state,
                statusvisible: true,
                status: action.value
            };
        case 'SHOW_DIALOG':
            return {
                ...state,
                dialog: action.value
            };
        case 'DIALOG_HIDE':
            return {
                ...state,
                dialog: {
                    title: "",
                    contents: null,
                    open: false
                }
            };
    }
    return state
}
