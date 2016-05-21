// ObjectCard is a card that holds the properties shared between users/devices/Streams
// it is the main card that shows up when you query a user/device/stream, and allows
// you to expand, click to edit, etc.
// Since the specific components are different for the different object types, this card
// allows children which will display the properties however they want - it only handles shared properties.

import React, {Component, PropTypes} from 'react';

import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import storage from '../storage';

import AvatarIcon from './AvatarIcon';

import '../util';

class ObjectCard extends Component {
    static propTypes = {
        expanded: PropTypes.bool.isRequired,
        onEditClick: PropTypes.func.isRequired,
        onExpandClick: PropTypes.func.isRequired,
        object: PropTypes.object.isRequired,
        path: PropTypes.string.isRequired,
        style: PropTypes.object
    }

    render() {
        let obj = this.props.object;
        let nickname = obj.name.capitalizeFirstLetter();
        if (obj.nickname !== undefined && obj.nickname != "") {
            nickname = obj.nickname;
        }
        let showedit = (obj['user_editable'] === undefined || obj['user_editable']);
        return (
            <Card style={this.props.style} onExpandChange={this.props.onExpandClick} expanded={this.props.expanded}>
                <CardHeader title={nickname} subtitle={this.props.path} showExpandableButton={true} avatar={< AvatarIcon name = {
                    obj.name
                }
                iconsrc = {
                    obj.icon
                } />}>
                    {this.props.expanded
                        ? (
                            <div style={{
                                float: "right",
                                marginRight: 35,
                                marginTop: "-5px",
                                marginLeft: "-100px"
                            }}>
                                {showedit
                                    ? (
                                        <IconButton onTouchTap={() => this.props.onEditClick(true)} tooltip="edit">
                                            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                                edit
                                            </FontIcon>
                                        </IconButton>
                                    )
                                    : null}

                                <IconButton onTouchTap= { () => storage.query(this.props.path) } tooltip="reload">
                                    <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                        refresh
                                    </FontIcon>
                                </IconButton>
                            </div>
                        )
                        : null}
                </CardHeader>
                <CardText expandable={true}>
                    {obj.description == ""
                        ? (null)
                        : (
                            <div style={{
                                color: "grey"
                            }}>{obj.description}</div>
                        )}
                    {this.props.children}
                </CardText>
            </Card>
        );
    }
}
export default ObjectCard;
