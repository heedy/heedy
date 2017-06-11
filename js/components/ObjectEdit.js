// The setup used for editing users/devices/streams. It includes all of the fields
// shared between users/devices/streams

import React, { Component } from "react";
import PropTypes from "prop-types";

import { Card, CardText, CardHeader, CardActions } from "material-ui/Card";
import Dialog from "material-ui/Dialog";
import FlatButton from "material-ui/FlatButton";
import TextField from "material-ui/TextField";
import Checkbox from "material-ui/Checkbox";

import NicknameEditor from "./edit/NicknameEditor";
import DescriptionEditor from "./edit/DescriptionEditor";
import IconEditor from "./edit/IconEditor";

import AvatarIcon from "./AvatarIcon";
import "../util";

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
  };

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
    this.setState({ dialogopen: false });
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
      <Card
        style={{
          textAlign: "left"
        }}
      >
        <CardHeader
          title={nickname}
          subtitle={this.props.path}
          avatar={
            <AvatarIcon
              name={obj.name}
              iconsrc={edits.icon !== undefined ? edits.icon : obj.icon}
            />
          }
        />
        <CardText>
          <NicknameEditor
            type={this.props.objectLabel}
            value={edits.nickname !== undefined ? edits.nickname : obj.nickname}
            onChange={this.props.callbacks.nicknameChange}
          />
          <DescriptionEditor
            type={this.props.objectLabel}
            value={
              edits.description !== undefined
                ? edits.description
                : obj.description
            }
            onChange={this.props.callbacks.descriptionChange}
          />
          <IconEditor
            type={this.props.objectLabel}
            value={edits.icon !== undefined ? edits.icon : obj.icon}
            onChange={this.props.callbacks.iconChange}
          />
          {this.props.children}
        </CardText>
        <CardActions>
          <FlatButton
            primary={true}
            label="Save"
            onTouchTap={() => this.save()}
          />
          <FlatButton label="Cancel" onTouchTap={this.props.onCancel} />
          <FlatButton
            label="Delete"
            style={{
              color: "red",
              float: "right"
            }}
            onTouchTap={() => this.setState({ dialogopen: true })}
          />
        </CardActions>
        <Dialog
          title={"Delete " + objCapital}
          actions={[
            <FlatButton
              label="Cancel"
              onTouchTap={() => this.setState({ dialogopen: false })}
              keyboardFocused={true}
            />,
            <FlatButton
              label="Delete"
              onTouchTap={() => this.dialogDeleteClick()}
            />
          ]}
          modal={false}
          open={this.state.dialogopen}
        >
          Are you sure you want to delete the {this.props.objectLabel}
          {" "}
          "{this.props.path}"?
        </Dialog>
      </Card>
    );
  }
}
export default ObjectEdit;
