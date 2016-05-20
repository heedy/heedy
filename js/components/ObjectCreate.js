// ObjectCreate sets up the creation of objects (user/device/stream) - it is the main view used
// as well as the final buttons

import React, {Component, PropTypes} from 'react';
import {Card, CardText, CardHeader, CardActions} from 'material-ui/Card';
import TextField from 'material-ui/TextField';
import Divider from 'material-ui/Divider';
import FlatButton from 'material-ui/FlatButton';

import AvatarIcon from './AvatarIcon';
import NicknameEditor from './NicknameEditor';
import DescriptionEditor from './DescriptionEditor';

import '../util';

class ObjectCreate extends Component {
    static propTypes = {
        state: PropTypes.object.isRequired,
        callbacks: PropTypes.object.isRequired,
        type: PropTypes.string.isRequired,
        parentPath: PropTypes.string.isRequired,
        onCancel: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired,
        required: PropTypes.element
    }

    render() {
        let state = this.props.state;
        let title = "Create a new " + this.props.type;
        let subtitle = this.props.parentPath + "/" + state.name;
        let callbacks = this.props.callbacks;
        return (
            <Card style={{
                textAlign: "left"
            }}>
                <CardHeader title={title} subtitle={subtitle} avatar={< AvatarIcon name = {
                    state.name == ""
                        ? "?"
                        : state.name
                }
                iconsrc = {
                    state.icon
                } />}/>
                <CardText>
                    <h2>Required:</h2>
                    <h3>Name</h3>
                    <p>A name for your {this.props.type}. Try to make it all lowercase without any spaces.</p>
                    <TextField hintText={"my" + this.props.type} floatingLabelText="Name" value={state.name} onChange={callbacks.nameChange}/><br/> {this.props.required}
                    <Divider/>
                    <NicknameEditor type={this.props.type} value={state.nickname} onChange={callbacks.nicknameChange}/>
                    <DescriptionEditor type={this.props.type} value={state.description} onChange={callbacks.descriptionChange}/> {this.props.children}
                </CardText>
                <CardActions>
                    <FlatButton primary={true} label="Create" onTouchTap={this.props.onSave}/>
                    <FlatButton label="Cancel" onTouchTap={this.props.onCancel}/>
                </CardActions>
            </Card>
        );
    }
}
export default ObjectCreate;
