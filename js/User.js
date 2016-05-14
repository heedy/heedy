import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import prettydate from 'pretty-date';

import {Card, CardText, CardHeader, CardActions} from 'material-ui/Card';
import {
    Table,
    TableBody,
    TableHeader,
    TableHeaderColumn,
    TableRow,
    TableRowColumn
} from 'material-ui/Table';
import Avatar from 'material-ui/Avatar';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import {GridList, GridTile} from 'material-ui/GridList';
import Subheader from 'material-ui/Subheader';

import ReactMarkdown from 'react-markdown';

import storage from './storage';

import Error from './Error';
import Loading from './Loading';

class User extends Component {
    static propTypes = {
        params: PropTypes.shape({user: PropTypes.string.isRequired}).isRequired
    };

    constructor(props) {
        super(props);
        this.state = {
            user: null,
            error: null
        };
        this.isquerying = false;
    }

    // Upon mounting, create a callback which will update the component state when the user is updated
    componentDidMount() {
        // https://stackoverflow.com/questions/1349404/generate-a-string-of-5-random-characters-in-javascript
        this.callbackID = Math.random().toString(36).substring(7);

        storage.addCallback(this.callbackID, (path, obj) => {
            // If the current user was updated, update the view
            if (path == this.props.params.user) {
                if (obj.ref !== undefined) {
                    this.setState({error: obj});
                } else {
                    this.setState({user: obj});
                }
            }
        });

        // componentWillReceiveProps is not called on mount for some reason... so let's do this
        this.componentWillReceiveProps(this.props);
    }
    componentWillUnmount() {
        storage.remCallback(this.callbackID);
    }

    // This is called whenever the user changes
    componentWillReceiveProps(nextProps) {
        if (this.isquerying) {
            return;
        }
        this.isquerying = true;

        // Now: The willReceiveProps is a lie, it is called even when the props are basically the same.
        // So we check if the one prop we care about is here
        if (this.state.user != null && this.state.user.name == nextProps.params.user) {
            return;
        }

        // Set the user to null, since the old user should not longer be displayed
        this.setState({user: null, error: null});

        // Get the user from cache - this allows the app to feel fast on slow internet,
        // and enables it to work in offline mode
        storage.get(nextProps.params.user).then((response) => {
            this.isquerying = false;
            // Set the user to the cached value
            if (response != null) {
                if (response.ref !== undefined) {
                    this.setState({error: response});
                } else if (response.name !== undefined) {
                    this.setState({user: response});

                    // If the user was recently queried, don't query it again needlessly
                    if (response.timestamp > Date.now() - 1000 * 10) {
                        return;
                    }
                }
            }

            storage.query(nextProps.params.user).catch((err) => console.log(err));
        }).catch((err) => {
            this.isquerying = false;
            console.log(err);
        });
    }

    render() {
        if (this.state.error != null) {
            // There was an error
            return (<Error err={this.state.error}/>);
        }
        if (this.state.user == null) {
            // The user is currently being queried
            return (<Loading/>);
        }
        return (
            <div>
                <Card style={{
                    textAlign: "left"
                }}>
                    <CardHeader title={this.state.user.nickname != ""
                        ? this.state.user.nickname
                        : this.state.user.name} subtitle={this.state.user.name} actAsExpander={true} showExpandableButton={true} avatar={< Avatar > U < /Avatar>}>
                        <IconButton style={{
                            float: "right",
                            marginRight: 30,
                            marginTop: -5
                        }} onTouchTap={this.props.hamburgerClick}>
                            <FontIcon className="material-icons">
                                edit
                            </FontIcon>
                        </IconButton>
                    </CardHeader>
                    <CardText expandable={true}>
                        {this.state.user.description == ""
                            ? (null)
                            : (
                                <div style={{
                                    color: "grey"
                                }}><ReactMarkdown escapeHtml={true} source={this.state.user.description}/></div>
                            )}
                        <Table selectable={false}>
                            <TableHeader enableSelectAll={false} displaySelectAll={false} adjustForCheckbox={false}>
                                <TableRow>
                                    <TableHeaderColumn>Email</TableHeaderColumn>
                                    <TableHeaderColumn>Public</TableHeaderColumn>
                                    <TableHeaderColumn>Role</TableHeaderColumn>
                                </TableRow>
                            </TableHeader>
                            <TableBody displayRowCheckbox={false}>
                                <TableRow>
                                    <TableRowColumn>{this.state.user.email}</TableRowColumn>
                                    <TableRowColumn>{this.state.user.public
                                            ? "true"
                                            : "false"}</TableRowColumn>
                                    <TableRowColumn>{this.state.user.role}</TableRowColumn>
                                </TableRow>
                            </TableBody>
                        </Table>

                    </CardText>
                </Card>
                <Subheader style={{
                    marginTop: 20
                }}>Devices</Subheader>

            </div>
        );
    }
}
export default User;
