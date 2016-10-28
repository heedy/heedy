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
import IconMenu from 'material-ui/IconMenu';
import MenuItem from 'material-ui/MenuItem';
import MoreVertIcon from 'material-ui/svg-icons/navigation/more-vert';
import NavigationClose from 'material-ui/svg-icons/navigation/close';
import {getSearchState} from '../reducers/search';

// setSearchText is called whenever the user changes the search box text. All actions happen through setSearchText
import {setSearchText} from '../actions'

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
        search: React.PropTypes.object.isRequired,
        hamburgerClick: React.PropTypes.func,
        searchTextChanged: React.PropTypes.func,
        router: React.PropTypes.object
    };

    render() {

        // The search bar can have
        let search = this.props.search;

        return (
            <Toolbar style={{
                height: `${spacing.desktopKeylineIncrement}px`,
                background: "#009e42",
                boxShadow: "0px 2px 5px #888888",
                position: "fixed",
                width: "100%",
                top: "0px",
                zIndex: 999
            }}>
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
                        {search.icon}
                    </FontIcon>
                    <TextField hintText={search.hint} style={{
                        paddingLeft: "10px",
                        fontWeight: "bold"
                    }} inputStyle={{
                        color: "white"
                    }} fullWidth={true} underlineShow={false} value={search.text} onChange={this.props.searchTextChanged}/> {search.text == ""
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
                <ToolbarGroup style={{
                    marginTop: "7px",
                    marginLeft: "10px"
                }}>
                    <IconMenu iconButtonElement={< IconButton > <MoreVertIcon/> < /IconButton>} anchorOrigin={{
                        horizontal: 'right',
                        vertical: 'top'
                    }} targetOrigin={{
                        horizontal: 'left',
                        vertical: 'bottom'
                    }}>
                        {this.props.menu.map((link) => {
                            return (<MenuItem key={link.title} primaryText={link.title} leftIcon={< FontIcon className = "material-icons" > {
                                link.icon
                            } < /FontIcon>} onTouchTap={() => link.action(this.props.dispatch)}/>)
                        })}
                    </IconMenu>
                </ToolbarGroup>
            </Toolbar>
        );
    }
}

export default connect((state) => ({search: getSearchState(state), menu: state.site.dropdownMenu}), (dispatch, props) => ({
    searchTextChanged: (e, txt) => dispatch(setSearchText(txt)),
    dispatch: dispatch
}))(TopBar);
