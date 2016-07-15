/*
 The TopBar is the bar shown at the top of the app, and it includes a search box.
 If on mobile, it also shows the hamburger menu (which activates the navigation). This component
 is added to the app in Theme.js
*/

import React, {Component} from 'react';
import {connect} from 'react-redux';

import {spacing} from 'material-ui/styles';
import FontIcon from 'material-ui/FontIcon';
import {Toolbar, ToolbarGroup, ToolbarSeparator, ToolbarTitle} from 'material-ui/Toolbar';
import TextField from 'material-ui/TextField';
import IconButton from 'material-ui/IconButton';
import NavigationClose from 'material-ui/svg-icons/navigation/close';

// setSearchText is called whenever the user changes the search box text. All actions happen through setSearchText
import {setSearchText} from './actions'

const styles = {
    searchbar: {
        //marginLeft: "10px",
        marginRight: "10px",
        marginTop: "10px",
        marginBottom: "10px",
        background: "#00b34a",
        width: "100%",
        borderRadius: "5px"
    }
};

class TopBar extends Component {
    static propTypes = {
        navDocked: React.PropTypes.bool.isRequired,
        searchText: React.PropTypes.string.isRequired,
        hamburgerClick: React.PropTypes.func,
        searchTextChanged: React.PropTypes.func
    };

    render() {
        return (
            <Toolbar style={{
                height: `${spacing.desktopKeylineIncrement}px`,
                background: "#009e42",
                boxShadow: "0px 2px 5px #888888",
                position: "fixed",
                width: "100%",
                top: "0px",
                zIndex: 999
            }} zDepth={50}>
                {this.props.navDocked
                    ? null
                    : (
                        <ToolbarGroup firstChild={true}>
                            <IconButton style={{
                                marginTop: "7px",
                                paddingLeft: "20px",
                                paddingRight: "40px"
                            }} onTouchTap={this.props.hamburgerClick}>
                                <FontIcon className="material-icons" color="#00662a" style={{
                                    fontSize: "80px"
                                }}>
                                    menu
                                </FontIcon>
                            </IconButton>
                        </ToolbarGroup>
                    )}
                <ToolbarGroup firstChild={this.props.navDocked} style={Object.assign({}, styles.searchbar, this.props.navDocked
                    ? {
                        marginLeft: "266px"
                    }
                    : {
                        marginLeft: "10px"
                    })}>
                    <FontIcon className="material-icons" style={{
                        marginTop: "-5px"
                    }}>
                        search
                    </FontIcon>
                    <TextField hintText="Search" style={{
                        paddingLeft: "10px",
                        fontWeight: "bold"
                    }} inputStyle={{
                        color: "white"
                    }} fullWidth={true} underlineShow={false} value={this.props.searchText} onChange={this.props.searchTextChanged}/> {this.props.searchText == ""
                        ? null
                        : (
                            <FontIcon className="material-icons" style={{
                                marginTop: "-5px",
                                paddingRight: "10px"
                            }} onTouchTap={() => this.props.searchTextChanged(null, "")}>
                                close
                            </FontIcon>
                        )}
                </ToolbarGroup>
            </Toolbar>
        );
    }
}

export default connect((state) => ({searchText: state.query.queryText}), (dispatch) => ({
    searchTextChanged: (e, txt) => dispatch(setSearchText(txt))
}))(TopBar);
