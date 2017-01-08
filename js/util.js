// basic utilities that don't fit in anywhere else

// react-router uses an old version of history that doesn't have a way to extract location... so we have to listen
// to the locations, and export current location

// Allows us to get the redux store as "app"
export var app = null;
export function setApp(store) {
    app = store;
}


// https://stackoverflow.com/questions/1026069/capitalize-the-first-letter-of-string-in-javascript
String.prototype.capitalizeFirstLetter = function () {
    return this.charAt(0).toUpperCase() + this.slice(1);
}

// setTitle sets the window title.
export function setTitle(txt) {
    if (txt == "" || txt === undefined || txt === null) {
        document.title = "ConnectorDB";
    } else {
        document.title = txt + " - ConnectorDB";
    }
}

// Strips the beginning / and end / from the path
export function getCurrentPath() {
    let p = window.location.pathname.substring(1, location.pathname.length);
    if (p.endsWith("/")) {
        p = p.substring(0, p.length - 1);
    }
    return p;
}

function keepElement(t, elem) {
    if (elem.name.toLowerCase().indexOf(t) != -1) {
        return true;
    }
    if (elem.nickname.toLowerCase().indexOf(t) != -1) {
        return true;
    }
    if (elem.description.toLowerCase().indexOf(t) != -1) {
        return true;
    }
    return false;
}

// This filters by keepElement - makes search really easy
// http://stackoverflow.com/a/37616104
export function objectFilter(text, obj) {
    let t = text.trim().toLowerCase();
    if (t.length == 0) {
        return obj;
    }

    return Object.keys(obj).filter(key => keepElement(t, obj[key])).reduce((res, key) => (res[key] = obj[key], res), {});
}

// Uses a WebRTC hack to get the local IP. It is used when we only hae "localhost" as a server identifier
// https://github.com/diafygi/webrtc-ips
export function getIPs(callback) {
    var ip_dups = {};

    //compatibility for firefox and chrome
    var RTCPeerConnection = window.RTCPeerConnection
        || window.mozRTCPeerConnection
        || window.webkitRTCPeerConnection;
    var useWebKit = !!window.webkitRTCPeerConnection;

    //bypass naive webrtc blocking using an iframe
    if (!RTCPeerConnection) {
        //NOTE: you need to have an iframe in the page right above the script tag
        //
        //<iframe id="iframe" sandbox="allow-same-origin" style="display: none"></iframe>
        //<script>...getIPs called in here...
        //
        var win = iframe.contentWindow;
        RTCPeerConnection = win.RTCPeerConnection
            || win.mozRTCPeerConnection
            || win.webkitRTCPeerConnection;
        useWebKit = !!win.webkitRTCPeerConnection;
    }

    //minimal requirements for data connection
    var mediaConstraints = {
        optional: [{ RtpDataChannels: true }]
    };

    var servers = { iceServers: [{ urls: "stun:stun.services.mozilla.com" }] };

    //construct a new RTCPeerConnection
    var pc = new RTCPeerConnection(servers, mediaConstraints);

    function handleCandidate(candidate) {
        //match just the IP address
        var ip_regex = /([0-9]{1,3}(\.[0-9]{1,3}){3}|[a-f0-9]{1,4}(:[a-f0-9]{1,4}){7})/
        var ip_addr = ip_regex.exec(candidate)[1];

        //remove duplicates
        if (ip_dups[ip_addr] === undefined)
            callback(ip_addr);

        ip_dups[ip_addr] = true;
    }

    //listen for candidate events
    pc.onicecandidate = function (ice) {

        //skip non-candidate events
        if (ice.candidate)
            handleCandidate(ice.candidate.candidate);
    };

    //create a bogus data channel
    pc.createDataChannel("");

    //create an offer sdp
    pc.createOffer(function (result) {

        //trigger the stun server request
        pc.setLocalDescription(result, function () { }, function () { });

    }, function () { });

    //wait for a while to let everything done
    setTimeout(function () {
        //read candidate info from local description
        var lines = pc.localDescription.sdp.split('\n');

        lines.forEach(function (line) {
            if (line.indexOf('a=candidate:') === 0)
                handleCandidate(line);
        });
    }, 1000);
}