// The ObjectList is a list of Objects, where the object can be users, devices or streams - it only uses properties shared between
// all three to render them (except for Public - if the user/device is public, it adds a little social icon)

import React, {Component, PropTypes} from 'react';

import {Card, CardText, CardHeader} from 'material-ui/Card';
import {List, ListItem} from 'material-ui/List';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';
import Divider from 'material-ui/Divider';
import Avatar from 'material-ui/Avatar';

import '../util';
import AvatarIcon from './AvatarIcon';

class ObjectList extends Component {
    static propTypes = {
        objects: PropTypes.object.isRequired,
        addName: PropTypes.string.isRequired,
        onAddClick: PropTypes.func,
        onSelect: PropTypes.func,
        style: PropTypes.object
    }

    static defaultProps = {
        style: {},
        onSelect: () => {},
        onAddClick: () => {}
    }

    render() {
        let addName = this.props.addName;
        return (
            <Card style={this.props.style}>
                <List>
                    {Object.keys(this.props.objects).map((key) => {
                        let obj = this.props.objects[key];
                        let primaryText = (obj.nickname != ""
                            ? obj.nickname
                            : obj.name.capitalizeFirstLetter());

                        return (
                            <div key={key}>
                                <ListItem primaryText={primaryText} secondaryText={obj.description} leftAvatar={< AvatarIcon name = {
                                    obj.name
                                }
                                iconsrc = {
                                    obj.icon
                                } />} rightIcon={obj.public
                                    ? (
                                        <IconButton style={{
                                            paddingRight: "30px",
                                            marginTop: "0px"
                                        }} tooltip={"public"} disabled={true}>
                                            <FontIcon className="material-icons">group</FontIcon>
                                        </IconButton>
                                    )
                                    : undefined} onTouchTap={() => this.props.onSelect(key)}/>
                                <Divider inset={true}/>
                            </div>
                        );
                    })}
                    <ListItem primaryText={"Add " + addName.capitalizeFirstLetter()} secondaryText={"Create a new " + addName} onTouchTap={this.props.onAddClick} leftAvatar={< Avatar icon = { < FontIcon className = "material-icons" > add < /FontIcon>} / >}/>
                </List>
            </Card >
        );
    }
}
export default ObjectList;
