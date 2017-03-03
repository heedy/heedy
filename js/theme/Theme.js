/*
  The Theme component renders the main website (the navigation, notification box etc).
  The theme's children are the actual content pages.

  This component puts the entire theme together, including colors. It also sets up responsiveness - it
  queries the browser width to determine whether the navigation is to be docked or not (ie, whether the app is
  running on a phone or not).

  Furthermore, this component renders the search results page directly, and it renders the snackbar directly (the
  notification popup at bottom of screen).
*/

import React, { Component, PropTypes } from 'react';
import { connect } from 'react-redux';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import Snackbar from 'material-ui/Snackbar';
import Dialog from 'material-ui/Dialog';
import FlatButton from 'material-ui/FlatButton';

import { spacing } from 'material-ui/styles';
import withWidth, { MEDIUM, LARGE } from 'material-ui/utils/withWidth';

import Navigation from './Navigation';
import TopBar from './TopBar';

// muiTheme represents our color scheme for the material design UI
const muiTheme = getMuiTheme({
    palette: {
        primary1Color: "#005c9e",
        primary2Color: "#009e91",
        primary3Color: "#009e42"
    },
    zIndex: {
        tooltip: 9999
    }
});

const styles = {
    container: {
        textAlign: 'center',
        paddingLeft: 256,
        background: "#F0F0F0",
        paddingBottom: "20px"
    },
    containerFullWidth: {
        textAlign: 'center',
        background: "#F0F0F0",
        paddingBottom: "20px"
    },
    mainStyle: {
        marginTop: `${spacing.desktopKeylineIncrement + 20}px`,
        marginRight: "auto",
        marginLeft: "auto",
        maxWidth: "80%"
    },
    mainStyleFullWidth: {
        marginTop: `${spacing.desktopKeylineIncrement + 20}px`,
        marginRight: "10px",
        marginLeft: "10px"
    }
};

class Theme extends Component {
    static propTypes = {
        width: PropTypes.number.isRequired,
        location: PropTypes.object.isRequired,
        router: PropTypes.object
    };

    constructor(props) {
        super(props);
        this.state = {
            drawerOpen: false
        };
    }

    render() {
        let isNavigationDocked = this.props.width === MEDIUM || this.props.width === LARGE;
        let curloc = this.props.location.pathname + this.props.location.hash;
        return (
            <MuiThemeProvider muiTheme={muiTheme}>
                <div>
                    <Navigation docked={isNavigationDocked} selected={curloc.substring(1, curloc.length)} open={this.state.drawerOpen} onRequestChange={(open) => this.setState({ drawerOpen: open })} />
                    <TopBar navDocked={isNavigationDocked} router={this.props.router} hamburgerClick={() => this.setState({ drawerOpen: true })} />
                    <div style={isNavigationDocked
                        ? styles.container
                        : styles.containerFullWidth}>
                        <div style={isNavigationDocked
                            ? styles.mainStyle
                            : styles.mainStyleFullWidth}>
                            <div>{this.props.children}</div>
                            <div className="col-lg-12" style={{
                                textAlign: "center",
                                font: "light grey",
                                fontSize: "8px",
                                marginTop: "20px"
                            }}></div>
                        </div>
                    </div>
                    <Dialog title={this.props.dialog.title} onRequestClose={this.props.dialogClose} actions={[(<FlatButton label="OK" onTouchTap={() => this.props.dialogClose()} keyboardFocused={true} />)]} modal={false} open={this.props.dialog.open}>
                        {this.props.dialog.contents}
                    </Dialog>
                    <Snackbar autoHideDuration={4000} message={this.props.message} open={this.props.showmsg} onRequestClose={this.props.onMsgClose} />
                </div>
            </MuiThemeProvider>
        );
    }
}

export default connect((state) => ({ message: state.site.status, showmsg: state.site.statusvisible, dialog: state.site.dialog }), (dispatch) => ({
    onMsgClose: () => dispatch({ type: 'STATUS_HIDE' }),
    dialogClose: () => dispatch({ type: "DIALOG_HIDE" })

}))(withWidth()(Theme));
