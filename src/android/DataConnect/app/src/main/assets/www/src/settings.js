
var SettingsPage = React.createClass({

	handleLogout: function(val) {
		connector.setCredentials("","");
		app.setUsername("");
		app.setApiKey("");

		console.log("Logout");

		//Show the login screen
		app.render(<LoginForm />)
	},
	handleBack: function() {
		app.render(<MainPage />);
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
					<div className="toggle"><div className="toggle-handle"></div></div>
			  </li>
			  <li className="table-view-divider">Account</li>
			  <li className="table-view-cell" onClick={this.handleLogout}>Log Out</li>
			</ul>
			</div>
			</div>
		);
	}
});
