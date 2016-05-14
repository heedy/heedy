import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {Card, CardText, CardHeader, CardActions} from 'material-ui/Card';
import Avatar from 'material-ui/Avatar';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import ReactMarkdown from 'react-markdown';

import UserView from './UserView';
import UserEdit from './UserEdit';

import storage from '../storage';

class UserCard extends Component {
    constructor(props) {
        super(props);
        this.state = {
            expanded: false
        }
    }
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        editing: PropTypes.bool.isRequired,
        onEditClick: PropTypes.func
    }

    render() {
        let nickname = this.props.user.name;
        if (this.props.user.nickname !== undefined && this.props.user.nickname != "") {
            nickname = this.props.user.nickname;
        }

        return (
            <Card style={{
                textAlign: "left"
            }} onExpandChange={(cng) => this.setState({expanded: cng})}>
                <CardHeader title={nickname} subtitle={this.props.user.name} actAsExpander={true} showExpandableButton={true} avatar={< Avatar > U < /Avatar>}>
                    {(this.state.expanded && !this.props.editing)
                        ? (
                            <div style={{
                                float: "right",
                                marginRight: 35,
                                marginTop: "-5px"
                            }}>
                                <IconButton onTouchTap={this.props.onEditClick} tooltip="edit">
                                    <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                        edit
                                    </FontIcon>
                                </IconButton>
                                <IconButton onTouchTap= { () => storage.query(this.props.user.name) } tooltip="reload">
                                    <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                        refresh
                                    </FontIcon>
                                </IconButton>
                            </div>
                        )
                        : null}
                </CardHeader>
                <CardText expandable={true}>
                    {this.props.editing
                        ? (<UserEdit user={this.props.user}/>)
                        : (<UserView user={this.props.user}/>)}
                </CardText>
            </Card>
        );
    }
}
export default UserCard;
