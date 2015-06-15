var app = {
    id: document.getElementById("container"),
    device: false,

    getApiKey: function() {
        if (typeof(Storage) !=="undefined") {
            return localStorage.getItem("cdb_apikey")
        }
        console.error("Could not access local storage")
        return ""
    },
    getUsername: function() {
        if (typeof(Storage) !=="undefined") {
            return localStorage.getItem("cdb_username")
        }
        console.error("Could not access local storage")
        return ""
    },
    setApiKey: function(val) {
        if (typeof(Storage) !=="undefined") {
            localStorage.setItem("cdb_apikey",val)
        } else {
            console.error("Could not access local storage")
        }
    },
    setUsername: function(val) {
        if (typeof(Storage) !=="undefined") {
            localStorage.setItem("cdb_username",val)
        } else {
            console.error("Could not access local storage")
        }
    },

    // Application Constructor
    initialize: function() {
        this.bindEvents();

    },

    render: function(e) {
        React.render(e,app.id);
    },

    // Bind Event Listeners
    //
    // Bind any events that are required on startup. Common events are:
    // 'load', 'deviceready', 'offline', and 'online'.
    bindEvents: function() {
        document.addEventListener('deviceready', this.onDeviceReady, false);
    },

    onDeviceReady: function () {
        //Override HTML alert with native dialog
        if (navigator.notification) {
            window.alert = function (message) {
                navigator.notification.alert(message, null, "DataConnect", "OK");
            };
        }
        console.log("The device is ready.")
        app.device = true;
    },

};

app.initialize();
FastClick.attach(document.body);
