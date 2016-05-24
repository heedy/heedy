// MainToolbar is the toolbar shown on the main page, from which you can create new rating streams,
// and so forth
import React, {Component, PropTypes} from 'react';

import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import storage from '../storage';

class MainToolbar extends Component {
    static propTypes = {}

    render() {
        return (
            <Card style={this.props.style}>
                <CardHeader title={"Hi!"} subtitle={"Let's gather some data!"}>

                    <div style={{
                        float: "right",
                        marginTop: "-5px",
                        marginLeft: "-100px"
                    }}>
                        <IconButton onTouchTap= { () => storage.query(this.props.path) } tooltip="add stream">
                            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                add
                            </FontIcon>
                        </IconButton>
                        <IconButton onTouchTap={() => this.props.onEditClick(true)} tooltip="add rating">
                            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                star
                            </FontIcon>
                        </IconButton>
                        <IconButton onTouchTap= { () => storage.query(this.props.path) } tooltip="reload">
                            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                refresh
                            </FontIcon>
                        </IconButton>
                    </div>
                </CardHeader>
            </Card>
        );
    }
}
export default MainToolbar;
