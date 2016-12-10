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
