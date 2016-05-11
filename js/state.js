const InitialState = {
    // The context is set up during runtime, before the app is displayed. It contains
    // the data sent from ConnectorDB about the current user
    context: null,

    // navigation is displayed in the app's main nmenu
    navigation: [
        {
            title: "{self}",
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
    searchText: ""
};

export default InitialState;
