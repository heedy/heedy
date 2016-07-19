// MainToolbar is the toolbar shown on the main page, from which you can create new rating streams,
// and so forth
import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';
import {go} from '../actions';

import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import storage from '../storage';

class MainToolbar extends Component {
    static propTypes = {
        onAddClick: PropTypes.func.isRequired,
        onRatingClick: PropTypes.func.isRequired
    }

    render() {
        return (
            <Card style={this.props.style}>
                <CardHeader title={"Hi!"} subtitle={"Let's Gather Data!"}>

                    <div style={{
                        float: "right",
                        marginTop: "-5px",
                        marginLeft: "-300px"
                    }}>
                        <IconButton onTouchTap={this.props.onAddClick} tooltip="add stream">
                            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                add
                            </FontIcon>
                        </IconButton>
                        <IconButton onTouchTap={this.props.onRatingClick} tooltip="add rating">
                            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                star
                            </FontIcon>
                        </IconButton>
                        <IconButton onTouchTap={this.props.onLogClick} tooltip="add log (diary)">
                            <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                                library_books
                            </FontIcon>
                        </IconButton>
                        <IconButton onTouchTap= { () => storage.qls(this.props.user.name+"/"+this.props.device.name) } tooltip="reload">
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
export default connect(undefined, (dispatch, props) => ({

    onAddClick: () => dispatch(go(props.user.name + "/" + props.device.name + "#create")),
    onRatingClick: () => dispatch(go(props.user.name + "/" + props.device.name + "#create/rating.stars")),
    onLogClick: () => dispatch(go(props.user.name + "/" + props.device.name + "#create/log.diary"))
}))(MainToolbar);
