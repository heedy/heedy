function ConnectorDB(user, password, url) {
    password = password | ""
    this.url = url | "https://connectordb.com"

    this.authHeader = "Basic " + btoa(user+":"+password)

    if (password.length == 0) {
        this.authHeader = "Basic " + btoa(":" + user)
    }
}

ConnectorDB.prototype = {
    thisdevice: function () {
        return this.url
    }
}