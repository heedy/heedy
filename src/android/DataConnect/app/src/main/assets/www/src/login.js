
var LoginForm = React.createClass({
	componentDidMount: function() {
		this.userlog = "";
	},

	handleSubmit: function(e) {
		e.preventDefault();
		var uname = React.findDOMNode(this.refs.username).value;
		var pwd = React.findDOMNode(this.refs.password).value;

		/*
		app.connector = new ConnectorDB(uname,pwd);

		lf = this
		lf.addLog("Knock Knock",false);

		app.connector.readUser(uname).then(function (result) {
			lf.addLog("Who's there?",true);
        }).catch(function (error) {
            lf.addLog("*cricket*",true);
			lf.addLog("Are you connected to the internet?", false)
        });
		*/


		app.setUsername(uname);
		app.setApiKey(pwd);
		//connector.setCredentials(uname,pwd);
		//connector.sync();
		console.log("Login attempt");

		//Show the main screen
		app.render(<MainPage />)

	},

	addLog: function(txt,remote) {
		if (remote) {
			txt = "<i>"+ txt+"</i>";
		}
		this.userlog = this.userlog + "<br />" + txt;
		this.setState({
			userlog: this.userlog
		});
	},


	clearLog: function() {
		this.setState({
			userlog: ""
		});
	},




	getInitialState: function() {

		return {
			username: app.getUsername(),
			password: app.getApiKey(),
			userlog: ""
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
			<div>
			<header className="bar bar-nav">
			<h1 className="title" id="title">ConnectorDB Login</h1>
			</header>
			<div className="content">
			<center><img src="img/logo.png" className="loginImage" /></center>
			<form onSubmit={this.handleSubmit} >
			<input type="text" ref="username" placeholder="Username"
				value={this.state.username} onChange={this.handleUsername} required autofocus />
			<input type="password" ref="password" placeholder="Password"
				value={this.state.password} onChange={this.handlePassword} required />
			<button className="btn btn-positive btn-block" type="submit">Sign in</button>
			</form>
			<p className="content-padded loginlog">{this.state.userlog}</p>
			</div>
			</div>
		);
	}
});
