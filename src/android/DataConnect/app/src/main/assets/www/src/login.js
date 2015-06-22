
var LoginForm = React.createClass({
	componentDidMount: function() {
		this.userlog = "";
	},

	handleSubmit: function(e) {
		e.preventDefault();
		var uname = React.findDOMNode(this.refs.username).value.trim().toLowerCase();
		var pwd = React.findDOMNode(this.refs.password).value;
		var loginbtn = React.findDOMNode(this.refs.password);

		loginbtn.disabled= true;
		app.connector = new ConnectorDB(uname,pwd);

		lf = this
		lf.clearLog();
		lf.addLog("Knock Knock",false);

		app.connector.readDevice(uname,"user").then(function (result) {
			lf.addLog("Who's there?",true);

			//Now log in using the API key of the user device
			pwd = result.apikey;
			app.connector = new ConnectorDB(uname+"/user",pwd);
			lf.addLog(uname+"'s phone, "+device.model+"!",false);
			app.connector.readDevice(uname,device.model).then(function(res) {
				lf.addLog("I know you! Come right in!",true);
				connector.setCredentials(uname+"/"+device.model,res.apikey);
				app.setUsername(uname);
				app.setApiKey(pwd);
				app.render(<MainPage />);
			}).catch(function(res) {
				if (res.status==401) {
					lf.addLog("Uhh... I don't know you.",true);
					lf.addLog("There was a problem setting up the phone.",false);
					loginbtn.disabled=false;
				} else if (res.status==404) {
					lf.addLog("Ooooh, shiny!",true);
					lf.addLog("Can I come in?",false);
					app.connector.createDevice(uname,device.model).then(function(res) {
						lf.addLog("Yes! Welcome!",true);
						connector.setCredentials(uname+"/"+device.model,res.apikey);
						app.setUsername(uname);
						app.setApiKey(pwd);
						app.render(<MainPage />);
					}).catch(function(res) {
						lf.addLog("No! "+res.result,true);
						lf.addLog("Looks like the server is being annoying...",false);
						loginbtn.disabled=false;
					});
				}
			});
        }).catch(function (req) {
			console.log("Connection error:"+req);
			if (req==null) {
				lf.addLog("*cricket*",true);
				lf.addLog("Are you connected to the internet?", false)
			} else if (req.status==401) {
				lf.addLog("Who's there?",true);
				lf.addLog(uname+"!",false);
				lf.addLog("Get off my lawn, "+uname+"!",true);
				lf.addLog("Looks like the username or password is wrong...",false);
			} else {
	            lf.addLog(req.status+": "+req.response,true);
				lf.addLog("It looks like the server is drunk...", false)
			}
			loginbtn.disabled=false;
        }).done();



		//app.setUsername(uname);
		//app.setApiKey(pwd);
		//connector.setCredentials(uname,pwd);
		//connector.sync();
		console.log("Login attempt");

		//Show the main screen
		//app.render(<MainPage />)

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
		this.userlog = "";
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
			<button className="btn btn-positive btn-block" ref="submit" type="submit">Sign in</button>
			</form>
			<p className="content-padded loginlog"><span dangerouslySetInnerHTML={{__html: this.state.userlog}} /></p>
			</div>
			</div>
		);
	}
});
