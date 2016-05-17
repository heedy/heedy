// basic utilities that don't fit in anywhere else


// https://stackoverflow.com/questions/1026069/capitalize-the-first-letter-of-string-in-javascript
String.prototype.capitalizeFirstLetter = function() {
    return this.charAt(0).toUpperCase() + this.slice(1);
}


export function setTitle(txt) {
    if (txt == "" || txt === undefined || txt === null) {
        document.title = "ConnectorDB";
    } else {
        document.title = txt + " - ConnectorDB";
    }
}
