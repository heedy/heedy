cordova.define("com.connectordb.dataconnect.dataconnector", function(require, exports, module) {
	var exec = require("cordova/exec");

	module.exports = {
		setCredentials: function(devicename,apikey) {
			exec(null,null,"DataConnectorPlugin","setcred",[devicename,apikey]);
		},
		cachesize: function(successCallback) {
		    exec(successCallback,null,"DataConnectorPlugin","getcachesize",[]);
		}
	};
});
