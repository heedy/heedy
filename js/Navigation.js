import React from 'react';
import {connect} from 'react-redux';

import {spacing} from 'material-ui/styles';
import Drawer from 'material-ui/Drawer';
import {List, ListItem, MakeSelectable} from 'material-ui/List';
import FontIcon from 'material-ui/FontIcon';

// This is directly from https://github.com/callemall/material-ui/blob/master/docs/src/app/components/AppNavDrawer.js
const SelectableList = MakeSelectable(List);

import {showPage} from './actions'

// styles are all of the colors and sizes for the underlying website theme
const styles = {
    // logo requires a manual div to be displayed correctly, so we set up all of the
    // coloring and box shadow to imitate the toolbar
    logo: {
        lineHeight: `${spacing.desktopKeylineIncrement}px`,
        background: "#009e42",
        paddingLeft: spacing.desktopGutter,
        marginBottom: 8,
        boxShadow: "0px 2px 5px #888888"
    },

    // These styles pertain to the main navigation menu
    menuInnerDivStyle: {
        paddingLeft: "50px",
        paddingTop: "0px",
        paddingBottom: "0px"
    },
    menuStyle: {
        paddingTop: "5px"
    },

    menuSubText: {
        fontSize: ".8em",
        color: "grey",
        lineHeight: ".8em",
        paddingBottom: "15px"
    },
    menuMainText: {
        lineHeight: "2em"
    }
};

class Navigation extends React.Component {
    static propTypes = {
        docked: React.PropTypes.bool.isRequired,
        open: React.PropTypes.bool.isRequired,
        links: React.PropTypes.arrayOf(React.PropTypes.object),
        selected: React.PropTypes.string.isRequired,
        onClick: React.PropTypes.func.isRequired,
        onRequestChange: React.PropTypes.func
    };

    // The navigation does not close itself when in mobile mode when clicked
    // we therefore manually close it on click
    onClick(e, v) {
        this.props.onClick(e, v);
        if (!this.props.docked) {
            this.props.onRequestChange(false);
        }

    }

    render() {
        return (
            <Drawer docked={this.props.docked} open={this.props.docked
                ? true
                : this.props.open} onRequestChange={this.props.onRequestChange}>
                <div style={styles.logo}>
                    <img src={SiteURL + "/app/title_logo_light.png"} style={{
                        height: "24px"
                    }}/>
                </div>
                <SelectableList value={this.props.selected} onChange={(e, v) => this.onClick(e, v)}>
                    {this.props.links.map((link) => (
                        <ListItem key={link.page} value={link.page} focusState={link.focused
                            ? 'focused'
                            : 'none'} leftIcon={< FontIcon className = "material-icons" style = {{color: "black"}} > {
                            link.icon
                        } < /FontIcon>} innerDivStyle={styles.menuInnerDivStyle} style={styles.menuStyle}>
                            <div>
                                <div style={styles.menuMainText}>{link.title}</div>
                                {link.subtitle == ""
                                    ? null
                                    : (
                                        <div style={styles.menuSubText}>
                                            {link.subtitle}
                                        </div>
                                    )}
                            </div>
                        </ListItem>
                    ))}
                </SelectableList>
            </Drawer>
        );
    }
}

export default connect((state) => ({links: state.app.navigation}), (dispatch) => ({
    onClick: (e, id) => {
        dispatch(showPage(id));
    }
}))(Navigation);
