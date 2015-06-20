var MainPage = React.createClass({

	getInitialState: function() {
		return {
			username: app.getUsername(),
			password: app.getApiKey(),
			cachelength: 0
		};
	},

	componentDidMount: function() {
		cupdater = this.cacheUpdate
		this.cachetimer = setInterval(function() {
			connector.cachesize(function(v) {
				cupdater(v);
			});
		},10000);
		if (!app.device) {
			app.deviceCallback = function() {
				connector.cachesize(cupdater);
			}
		} else {
			connector.cachesize(cupdater);
		}
	},

	componentWillUnmount: function() {
		clearInterval(this.cachetimer);
	},

	cacheUpdate: function(val) {
		this.setState({
			cachelength: val
		});
	},

	handleLogout: function(val) {
		connector.setCredentials("","");
		app.setUsername("");
		app.setApiKey("");

		console.log("Logout");

		//Show the login screen
		app.render(<LoginForm />)
	},

	handleSync: function(val) {
		console.log("Syncing");

		connector.sync();
	},


	render: function() {
		return (
			<div>
			<h1>{this.state.username}</h1>
			<p>There are {this.state.cachelength} datapoints in cache.</p>
			<a className="btn btn-lg btn-primary btn-block" onClick={this.handleSync}>Sync</a>
			<a className="btn btn-lg btn-primary btn-block" onClick={this.handleLogout}>Log Out</a>
			</div>
		);
	}
});
