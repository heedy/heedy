
var SettingsPage = React.createClass({

	componentDidMount: function() {
		document.querySelector("#bgsynctoggle").addEventListener("toggle",this.handleDataBG);
	},

	componentWillUnmount: function() {
		document.querySelector("#bgsynctoggle").removeEventListener("toggle",this.handleDataBG)
	},

	getInitialState: function() {
		return {
			databg: parseFloat(localStorage.getItem("settings_bgsync")) || -1
		};
	},

	handleLogout: function(val) {
		//Turn off background sync before logging out
		this.setBGSync(0);
		connector.setCredentials("","");
		app.setUsername("");
		app.setApiKey("");

		console.log("Logout");

		//Show the login screen
		app.render(<LoginForm />);
	},
	handleBack: function() {
		app.render(<MainPage />);
	},

	handleDataBG: function(e) {
		//If it was toggled off, set negative number. If toggled on, send the time period to sync with
		if (e.detail.isActive) {
			this.setBGSync(60*60);	//Once an hour
		} else {
			this.setBGSync(0);
		}
	},
	setBGSync: function(t) {
		localStorage.setItem("settings_bgsync",t);
		this.setState(
			{
				databg: t
			}
		);
		connector.setSync(t);
	},

	clearHandler: function() {
		connector.clear();
	},
	render: function() {
		return (
			<div>
			<header className="bar bar-nav bar-colored">
			<a className="icon icon-left-nav pull-left icon-nav" onClick={this.handleBack}></a>
			<h1 className="title" id="title">DataConnect Settings</h1>
			</header>
			<div className="content">
			<ul className="table-view">
			  <li className="table-view-divider">Data</li>
			  <li className="table-view-cell">Background Sync
					<div className={this.state.databg > 0? "toggle active": "toggle"} id="bgsynctoggle"><div className="toggle-handle"></div></div>
			  </li>
			<li className="table-view-cell" onClick={this.clearHandler}>Clear Cache</li>
			  <li className="table-view-divider">Account</li>
			  <li className="table-view-cell" onClick={this.handleLogout}>Log Out</li>
			</ul>
			</div>
			</div>
		);
	}
});
