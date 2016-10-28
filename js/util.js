// basic utilities that don't fit in anywhere else

// react-router uses an old version of history that doesn't have a way to extract location... so we have to listen
// to the locations, and export current location
import {browserHistory} from 'react-router';
export var location = {};
function updateLocation(loc) {
    location = loc;
}
browserHistory.listen(updateLocation);

// https://stackoverflow.com/questions/1026069/capitalize-the-first-letter-of-string-in-javascript
String.prototype.capitalizeFirstLetter = function() {
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
