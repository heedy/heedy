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

const styles = {
    headers: {
        color: "grey"
    }
};

class ObjectCreate extends Component {
    static propTypes = {
        state: PropTypes.object.isRequired,
        callbacks: PropTypes.object.isRequired,
        type: PropTypes.string.isRequired,
        parentPath: PropTypes.string.isRequired,
        onCancel: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired,
        required: PropTypes.element,
        advanced: PropTypes.func,
        header: PropTypes.string
    }

    static defaultProps = {
        advanced: null,
        header: ""
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
                    {this.props.header != ""
                        ? <p>{this.props.header}</p>
                        : null}
                    <h2 style={styles.headers}>Required:</h2>
                    <h3>Name</h3>
                    <p>A name for your {this.props.type}. Try to make it all lowercase without any spaces.</p>
                    <TextField hintText={"my" + this.props.type} floatingLabelText="Name" style={{
                        marginTop: "-25px"
                    }} value={state.name} onChange={callbacks.nameChange}/><br/> {this.props.required}
                    <Divider style={{
                        marginTop: "20px"
                    }}/>
                    <h2 style={styles.headers}>Optional:</h2>
                    <NicknameEditor type={this.props.type} value={state.nickname} onChange={callbacks.nicknameChange}/>
                    <DescriptionEditor type={this.props.type} value={state.description} onChange={callbacks.descriptionChange}/> {this.props.children}
                </CardText>
                <CardActions>
                    <FlatButton primary={true} label="Create" onTouchTap={this.props.onSave}/>
                    <FlatButton label="Cancel" onTouchTap={this.props.onCancel}/> {this.props.advanced != null
                        ? (<FlatButton label="Advanced" secondary={true} style={{
                            float: "right"
                        }} onTouchTap={this.props.advanced}/>)
                        : null}

                </CardActions>
            </Card>
        );
    }
}
export default ObjectCreate;
