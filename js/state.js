// This file contains the initial states used for setting up the state in redux

export const UserPageInitialState = {
    // whether currently editing the user
    editing: false,
    // Whether the user card is expanded
    expanded: false
};

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
                description: "has all permissions that the owning user has"
            }
        }
    },

    // navigation is displayed in the app's main nmenu
    navigation: [
        {
            title: "Progress Log",
            subtitle: "Manually insert data",
            icon: "star",
            page: "/"
        }, {
            title: "Profile",
            subtitle: "View your devices",
            icon: "face",
            page: "/{self}"
        }, {
            title: "Log Out",
            subtitle: "Exit your session",
            icon: "power_settings_new",
            page: "/logout"
        }
    ],

    // The text displayed in the search box
    searchText: "",

    // The currently logged in user and device. This is set up immediately on app start.
    // even before the app is rendered. Note that these are NOT updated along with
    // the app storage - this is the initial user and device
    thisUser: "",
    thisDevice: "",

    // The URL of the website, also available as global variable "SiteURL". This is set up
    // from the context on app load
    siteURL: "",

    // Page states are kept for every user/device/stream visited in this session.
    // This allows for back-and-forth between pages without losing your place!
    userpage: {},
    devicepage: {},
    streampage: {}
};

export default InitialState;

// get the user page from the state - the state might not have this
// particular page initialized, meaning that it wasn't acted upon
export function getUserState(user, state) {
    return (state.app.userpage[user] !== undefined
        ? state.app.userpage[user]
        : UserPageInitialState);
}
