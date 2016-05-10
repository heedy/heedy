import React, {Component, PropTypes} from 'react';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import {spacing} from 'material-ui/styles';
import withWidth, {MEDIUM, LARGE} from 'material-ui/utils/withWidth';

import Navigation from './Navigation'
import TopBar from './TopBar'

// muiTheme represents our color scheme for the material design UI
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
    containerFullWidth: {
        textAlign: 'center',
        background: "#F0F0F0"
    }
};

class Theme extends Component {
    static propTypes = {
        width: PropTypes.number.isRequired
    };

    constructor(props) {
        super(props);
        this.state = {
            drawerOpen: false
        };
    }

    render() {
        let isNavigationDocked = this.props.width === MEDIUM || this.props.width === LARGE;
        return (
            <MuiThemeProvider muiTheme={muiTheme}>
                <div>
                    <Navigation docked={isNavigationDocked} open={this.state.drawerOpen} onRequestChange={(open) => this.setState({drawerOpen: open})} links={[
                        {
                            title: "Hello Sir",
                            subtitle: "hi",
                            icon: "face",
                            value: "hi"
                        }, {
                            title: "Log Out",
                            subtitle: "Exit your session",
                            icon: "power_settings_new",
                            value: "hi2"
                        }
                    ]} onClick={(val) => console.log(val)} selected={""}/>
                    <TopBar navDocked={isNavigationDocked} searchText="" hamburgerClick={() => this.setState({drawerOpen: true})}/>
                    <div style={isNavigationDocked
                        ? styles.container
                        : styles.containerFullWidth}></div>
                </div>
            </MuiThemeProvider>
        );
    }
}

export default withWidth()(Theme);
