var StarRating = React.createClass({
	starnum: 10,

	getInitialState: function() {
		return {
			checkval: parseInt(localStorage.getItem(this.props.sname+"_rating")) || 0
		};
	},

	saveValue: function(num) {
		localStorage.setItem(this.props.sname+"_rating",num);
		this.setState({
			checkval: num
		});

		//Now attempt to connect to the server to set the value
		mythis = this;
		app.connector.insertStream(app.getUsername(),"user",mythis.props.sname,num).then(function(res) {
			console.log(mythis.props.sname+": Successfully inserted rating.")
		}).catch(function(err) {
			if (err.status >= 400) {
				console.log(mythis.props.sname+": Error inserting - attempting create.");
				app.connector.createStream(app.getUsername(),"user",mythis.props.sname,{type: "number",maximum: 10,minimum: 0}).then(function(res) {
					console.log(mythis.props.sname+": Create stream succeeded. Trying insert.");
					app.connector.insertStream(app.getUsername(),"user",mythis.props.sname,num).then(function(res) {
						console.log(mythis.props.sname+": Successfully inserted rating.")
					}).catch(function(err) {
						alert("Failed to save "+mythis.props.sname+" with error: "+err.status + " "+err.response);
					});
				}).catch(function(err) {
					alert("Failed to save "+mythis.props.sname+" with error: "+err.status + " "+err.response);
				}).done();
			} else {
				console.log(mythis.props.sname+": "+err.status + " "+err.response);
				alert("Failed to save "+mythis.props.sname+" with error: "+err.result);
			}
		}).done();
	},

	starClick: function(num) {
		this.saveValue(num);
	},
	render: function() {
		rows = [];
		for (var i=this.starnum; i >0; i--) {
			rows.push((
				<input id={this.props.sname+"_rating"+i} type="radio" name={this.props.sname} value={i} checked={this.state.checkval == i} />
			));
			rows.push((<label for={this.props.sname+"_rating"+i} onClick={this.starClick.bind(this,i)}>{i}</label>));
		}
		return (
			<div>
			<h4>{this.props.name}</h4>
				<span className="starRating">
				{rows}
				</span>
			</div>
		)
	}
});

var RatingView = React.createClass({
	render: function() {
		return (
			<div className="card">
			<ul className="table-view">
			<li className="table-view-cell table-view-divider">How are you feeling right now?</li>
			<li className="table-view-cell"><StarRating name="Mood" sname="rating_mood" /></li>
			<li className="table-view-cell"><StarRating name="Productivity" sname="rating_productivity"/></li>
			<li className="table-view-cell"><StarRating name="Life Satisfaction" sname="rating_satisfaction"/></li>
			<li className="table-view-cell"><StarRating name="Progress Towards Goals" sname="rating_progress" /></li>
			</ul>
			</div>
		);
	}
});

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

	handleSync: function(val) {
		console.log("Syncing");

		connector.sync();
	},

	handleSettings: function() {
		app.render(<SettingsPage />);
	},


	render: function() {
		return (
			<div>
			<header className="bar bar-nav bar-colored">
			  <a className="icon icon-gear pull-right icon-nav" onClick={this.handleSettings}></a>
				<button className="btn pull-left btn-nav"  onClick={this.handleSync}>Sync<span className="badge badge-positive">{this.state.cachelength}</span></button>
			  <h1 className="title">{this.state.username}</h1>
			</header>

			<div className="content">
			<RatingView />
			</div>
			</div>
		);
	}
});
