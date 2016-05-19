// The setup used for editing users/devices/streams. It includes all of the fields
// shared between users/devices/streams

import React, {Component, PropTypes} from 'react';

import {Card, CardText, CardHeader, CardActions} from 'material-ui/Card';
import Dialog from 'material-ui/Dialog';
import FlatButton from 'material-ui/FlatButton';
import {RadioButton, RadioButtonGroup} from 'material-ui/RadioButton';
import TextField from 'material-ui/TextField';
import Checkbox from 'material-ui/Checkbox';

import NicknameEditor from './NicknameEditor';
import DescriptionEditor from './DescriptionEditor';

import AvatarIcon from './AvatarIcon';
import '../util';

class ObjectEdit extends Component {
    static propTypes = {
        object: PropTypes.object.isRequired,
        path: PropTypes.string.isRequired,
        style: PropTypes.object,
        state: PropTypes.object.isRequired,
        objectLabel: PropTypes.string.isRequired,
        callbacks: PropTypes.object.isRequired,
        onCancel: PropTypes.func.isRequired,
        onDelete: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired
    }

    constructor(props) {
        super(props);
        this.state = {
            dialogopen: false
        };
    }

    save() {
        this.props.onSave();
    }

    dialogDeleteClick() {
        this.setState({dialogopen: false});
        this.props.onDelete();
    }

    render() {
        let objCapital = this.props.objectLabel.capitalizeFirstLetter();
        let obj = this.props.object;
        let edits = this.props.state;
        let nickname = obj.name.capitalizeFirstLetter();
        if (obj.nickname !== undefined && obj.nickname != "") {
            nickname = obj.nickname;
        }
        if (edits.nickname !== undefined && edits.nickname != "") {
            nickname = edits.nickname;
        }
        return (
            <Card style={{
                textAlign: "left"
            }}>
                <CardHeader title={nickname} subtitle={this.props.path} avatar={< AvatarIcon name = {
                    obj.name
                }
                iconsrc = {
                    obj.icon
                } />}/>
                <CardText>
                    <NicknameEditor type={this.props.objectLabel} value={edits.nickname !== undefined
                        ? edits.nickname
                        : obj.nickname} onChange={this.props.callbacks.nicknameChange}/>
                    <DescriptionEditor type={this.props.objectLabel} value={edits.description !== undefined
                        ? edits.description
                        : obj.description} onChange={this.props.callbacks.descriptionChange}/> {this.props.children}
                </CardText>
                <CardActions>
                    <FlatButton primary={true} label="Save" onTouchTap={() => this.save()}/>
                    <FlatButton label=" Cancel" onTouchTap={this.props.onCancel}/>
                    <FlatButton label="Delete" style={{
                        color: "red",
                        float: "right"
                    }} onTouchTap={() => this.setState({dialogopen: true})}/>
                </CardActions>
                <Dialog title={"Delete " + objCapital} actions={[(<FlatButton label="Cancel" onTouchTap={() => this.setState({dialogopen: false})} keyboardFocused={true}/>), (<FlatButton label="Delete" onTouchTap={() => this.dialogDeleteClick()}/>)]} modal={false} open={this.state.dialogopen}>
                    Are you sure you want to delete the {this.props.objectLabel}
                    {" "}
                    "{this.props.path}"?
                </Dialog>
            </Card>
        );
    }
}
export default ObjectEdit;
