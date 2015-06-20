
var LoginForm = React.createClass({
	handleSubmit: function(e) {
		e.preventDefault();
		var uname = React.findDOMNode(this.refs.username).value;
		var pwd = React.findDOMNode(this.refs.password).value;

		app.setUsername(uname);
		app.setApiKey(pwd);
		connector.setCredentials(uname,pwd);
		//connector.sync();
		console.log("Login attempt");

		//Show the main screen
		app.render(<MainPage />)

	},

	getInitialState: function() {
		return {
			username: app.getUsername(),
			password: app.getApiKey()
		};
	},
	handleUsername: function(event) {
		this.setState({
			username: event.target.value
		});
	},
	handlePassword: function(event) {
		this.setState({
			password: event.target.value
		});
	},

	render: function() {
		return (
			<form className="form-signin" onSubmit={this.handleSubmit} >
			<h2 className="form-signin-heading">ConnectorDB sign in</h2>
			<label htmlFor="inputEmail" className="sr-only">Username</label>
			<input type="text" ref="username" className="form-control" placeholder="Username"
				value={this.state.username} onChange={this.handleUsername} required autofocus />
			<label htmlFor="inputPassword" className="sr-only">Password</label>
			<input type="password" ref="password" className="form-control" placeholder="Password"
				value={this.state.password} onChange={this.handlePassword} required />
			<button className="btn btn-lg btn-primary btn-block" type="submit">Sign in</button>
			</form>
		);
	}
});
