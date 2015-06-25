function saveCred(uname,pass) {
	localStorage.setItem("connector_dname",uname+"/web_browser");
	localStorage.setItem("connector_apikey",pass);
	disable();
	window.close();
}

function makeStream(cdb, uname,pass) {
	connector.readStream(uname,"web_browser","history").then(function (result) {
		saveCred(uname,pass);
	}).catch(function(res) {
		if (res.status==401) {
			console.error(res.response);
		} else if (res.status==404) {
			connector.createStream(uname,"web_browser","history",{type: "object",properties: {url: {type: "string" }, title: {type: "string"}}}).then(function(res) {
				saveCred(uname,pass);
			}).catch(function(res) {
				console.log(res.response);
			});
		}
	}).done();
}

function handleLogin(e) {
	e.preventDefault();
	var uname = document.getElementById("inputUser").value;
	var pwd = document.getElementById("inputPassword").value;

	document.getElementById("loginform").removeEventListener("submit",handleLogin);

	connector = new ConnectorDB(uname,pwd);
	connector.readDevice(uname,"user").then(function (result) {
		connector.readDevice(uname,"web_browser").then(function(res) {
			makeStream(connector,uname,res.apikey);
		}).catch(function(res) {
			if (res.status==401) {
				console.error(res.response);
			} else if (res.status==404) {
				connector.createDevice(uname,"web_browser").then(function(res) {
					makeStream(connector,uname,res.apikey);
				}).catch(function(res) {
					console.log(res.response);
				});
			}
		});
	}).catch(function (req) {
		console.log("Connection error:"+req);
	}).done();

	return false;
}

function handleLogout(e) {
	e.preventDefault();
	document.getElementById("loginform").removeEventListener("submit",handleLogout);
	localStorage.setItem("connector_dname","");
	localStorage.setItem("connector_apikey","");
	enable();
}

function disable() {
	document.getElementById("inputUser").disabled=true;
	document.getElementById("inputPassword").disabled=true;
	document.getElementById("signinbtn").innerHTML= "Log out";
	document.getElementById("loginform")
          .addEventListener("submit", handleLogout, false);
}

function enable() {
	document.getElementById("inputUser").disabled=false;
	document.getElementById("inputUser").value="";
	document.getElementById("inputPassword").disabled=false;
	document.getElementById("inputPassword").value=""
	document.getElementById("signinbtn").innerHTML = "Sign in";
	document.getElementById("loginform")
          .addEventListener("submit", handleLogin, false);
}

function isLoggedIn() {
	dname = localStorage.getItem("connector_dname") || "";
	return (dname.length > 1)
}

window.addEventListener("load", function()
{
	if (isLoggedIn()) {
		disable();
	} else {
		enable();
	}
}, false);
