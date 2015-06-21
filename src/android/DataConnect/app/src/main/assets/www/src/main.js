var StarRating = React.createClass({
	render: function() {
		return (
			<div>
			<h4>{this.props.name}</h4>
			<div className="rating">
			<span>☆</span><span>☆</span><span>☆</span><span>☆</span><span>☆</span>
			<span>☆</span><span>☆</span><span>☆</span><span>☆</span><span>☆</span>
			</div>
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
			<li className="table-view-cell"><StarRating name="Mood" /></li>
			<li className="table-view-cell"><StarRating name="Productivity" /></li>
			<li className="table-view-cell"><StarRating name="Life Satisfaction" /></li>
			<li className="table-view-cell"><StarRating name="Progress Towards Goals" /></li>
			</ul>
			</div>
		)
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
