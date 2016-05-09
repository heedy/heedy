/*
The theme represents the container and menu which is displayed on the website. The theme builds up
the navigation
*/

import React from 'react';
import RaisedButton from 'material-ui/RaisedButton';
import Dialog from 'material-ui/Dialog';
import {deepOrange500} from 'material-ui/styles/colors';
import FlatButton from 'material-ui/FlatButton';
import Toolbar from 'material-ui/Toolbar';
import Drawer from 'material-ui/Drawer';
import MenuItem from 'material-ui/MenuItem';
import Avatar from 'material-ui/Avatar';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import {spacing, typography, zIndex} from 'material-ui/styles';
import {grey400} from 'material-ui/styles/colors';
import FileFolder from 'material-ui/svg-icons/file/folder';
import ActionGrade from 'material-ui/svg-icons/action/grade';
import FontIcon from 'material-ui/FontIcon';
import {
    Card,
    CardActions,
    CardHeader,
    CardMedia,
    CardTitle,
    CardText
} from 'material-ui/Card';

const muiTheme = getMuiTheme({

    palette: {
        primary1Color: "#005c9e",
        primary2Color: "#009e91",
        primary3Color: "#009e42"
    }
});

const styles = {
    container: {
        textAlign: 'center',
        paddingLeft: 256,
        background: "#F0F0F0"
    },
    logo: {
        cursor: 'pointer',
        fontSize: 24,
        color: typography.textFullWhite,
        lineHeight: `${spacing.desktopKeylineIncrement}px`,
        fontWeight: typography.fontWeightLight,
        background: "#009e42",
        paddingLeft: spacing.desktopGutter,
        marginBottom: 8,
        boxShadow: "0px 2px 5px #888888"
    },
    subText: {
        fontSize: ".8em",
        color: "grey",
        lineHeight: ".8em",
        paddingBottom: "15px"
    },
    divStyle: {
        lineHeight: "2em"
    },
    innerDivStyle: {
        paddingLeft: "50px"
    },
    mainStyle: {
        paddingTop: "5px"
    },
    cardStyle: {
        marginTop: "20px",
        marginRight: "auto",
        marginLeft: "auto",
        maxWidth: "80%"
    }
};

export function siteRenderer(node) {
    return (
        <MuiThemeProvider muiTheme={muiTheme}>
            <div>
                <Drawer>
                    <div style={styles.logo}><img src="app/title_logo_light.png" style={{
            height: "24px"
        }}/></div>
                    <MenuItem leftIcon={< FontIcon className = "material-icons" style = {
                        {
                            color: "black"
                        }
                    } > face < /FontIcon>} innerDivStyle={styles.innerDivStyle} style={styles.mainStyle}>
                        <div>
                            <div style={styles.divStyle}>dkumor</div>
                            <div style={styles.subText}>
                                View your Devices
                            </div>
                        </div>
                    </MenuItem>
                    <MenuItem leftIcon={< FontIcon className = "material-icons" style = {
                        {
                            color: "black"
                        }
                    } > explore < /FontIcon>} innerDivStyle={styles.innerDivStyle} style={styles.mainStyle}>
                        <div>
                            <div style={styles.divStyle}>Explore</div>
                            <div style={styles.subText}>
                                Ask about your data
                            </div>
                        </div>
                    </MenuItem>
                    <MenuItem leftIcon={< FontIcon className = "material-icons" style = {
                        {
                            color: "black"
                        }
                    } > power_settings_new < /FontIcon>} innerDivStyle={styles.innerDivStyle} style={styles.mainStyle}>Log Off</MenuItem>
                </Drawer>
                <div style={styles.container}>
                    <Toolbar style={{
                        height: `${spacing.desktopKeylineIncrement}px`,
                        background: "#009e42",
                        boxShadow: "0px 2px 5px #888888"
                    }} zDepth={50}></Toolbar>
                    <Card style={styles.cardStyle}>
                        <CardText>
                            <h1 >
                                material - ui
                            </h1>
                            <h2 >
                                example project
                            </h2>
                            <RaisedButton label="Super Secret Password" primary={true} onTouchTap={function() {
                                alert("hi");
                            }}/>
                            <node/>
                        </CardText>
                    </Card>
                </div >
            </div >
        </MuiThemeProvider >
    );
}
