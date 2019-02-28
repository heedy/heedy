/*
 The TopBar is the bar shown at the top of the app, and it includes a search box.
 If on mobile, it also shows the hamburger menu (which activates the navigation). This component
 is added to the app in Theme.js
*/

import React, { Component } from "react";
import PropTypes from "prop-types";
import { connect } from "react-redux";

import { spacing } from "material-ui/styles";
import FontIcon from "material-ui/FontIcon";
import {
  Toolbar,
  ToolbarGroup,
  ToolbarSeparator,
  ToolbarTitle
} from "material-ui/Toolbar";
import AutoComplete from "material-ui/AutoComplete";
import IconButton from "material-ui/IconButton";
import IconMenu from "material-ui/IconMenu";
import MenuItem from "material-ui/MenuItem";
import MoreVertIcon from "material-ui/svg-icons/navigation/more-vert";
import NavigationClose from "material-ui/svg-icons/navigation/close";
import { getSearchState, getSearchActionContext } from "../reducers/search";

// setSearchText is called whenever the user changes the search box text. All actions happen through setSearchText
import { setSearchText } from "../actions";

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
    navDocked: PropTypes.bool.isRequired,
    search: PropTypes.object.isRequired,
    hamburgerClick: PropTypes.func,
    searchTextChanged: PropTypes.func,
    submit: PropTypes.func
  };

  keypress(txt, idx) {
    if (idx == -1) {
      this.props.submit();
    }
  }

  render() {
    // The search bar can have
    let search = this.props.search;

    let erropts = {
      color: "rgba(0,0,0,0.3)"
    };
    let erropts2 = {};
    if (search.error != "") {
      erropts = {
        color: "yellow"
      };
      erropts2 = {
        tooltip: search.error,
        tooltipPosition: "bottom-right",
        touch: true
      };
    }

    return (
      <Toolbar
        style={{
          height: `${spacing.desktopKeylineIncrement}px`,
          backgroundColor: "#009e42",
          boxShadow: "0px 2px 5px #888888",
          position: "fixed",
          width: "100%",
          top: "0px",
          zIndex: 999
        }}
      >
        {this.props.navDocked
          ? null
          : <ToolbarGroup firstChild={true}>
              <IconButton
                style={{
                  paddingLeft: "20px",
                  paddingRight: "40px"
                }}
                onTouchTap={this.props.hamburgerClick}
              >
                <FontIcon
                  className="material-icons"
                  color="#00662a"
                  style={{
                    fontSize: "80px"
                  }}
                >
                  menu
                </FontIcon>
              </IconButton>
            </ToolbarGroup>}
        <ToolbarGroup
          firstChild={this.props.navDocked}
          style={Object.assign(
            {},
            styles.searchbar,
            this.props.navDocked
              ? {
                  marginLeft: "266px"
                }
              : {
                  marginLeft: "10px"
                }
          )}
        >
          <IconButton {...erropts2}>
            <FontIcon
              {...erropts}
              className="material-icons"
              style={{
                marginTop: "-5px"
              }}
            >
              {search.error ? "error_outline" : search.icon}
            </FontIcon>
          </IconButton>
          <AutoComplete
            disabled={!search.enabled}
            hintText={search.hint}
            filter={AutoComplete.noFilter}
            textFieldStyle={{
              fontWeight: "bold"
            }}
            menuStyle={{
              background: "#009e42"
            }}
            listStyle={{
              color: "white"
            }}
            inputStyle={{
              color: "white"
            }}
            hintStyle={{
              overflow: "hidden",
              height: "25px",
              bottom: "10px"
            }}
            fullWidth={true}
            underlineShow={false}
            open={search.error != ""}
            searchText={search.text}
            dataSource={search.autocomplete}
            onUpdateInput={this.props.searchTextChanged}
            onNewRequest={this.keypress.bind(this)}
          />
          {" "}
          {search.text == ""
            ? null
            : <FontIcon
                className="material-icons"
                style={{
                  paddingRight: "10px"
                }}
                onTouchTap={() => this.props.searchTextChanged("", null)}
              >
                close
              </FontIcon>}

        </ToolbarGroup>
        <ToolbarGroup
          style={{
            marginLeft: "-10px",
            marginRight: "-20px"
          }}
        >
          <IconMenu
            iconButtonElement={<IconButton> <MoreVertIcon /> </IconButton>}
            anchorOrigin={{
              horizontal: "right",
              vertical: "bottom"
            }}
            targetOrigin={{
              horizontal: "right",
              vertical: "top"
            }}
          >
            {this.props.menu.map(link => {
              return (
                <MenuItem
                  key={link.title}
                  primaryText={link.title}
                  leftIcon={
                    <FontIcon className="material-icons">
                      {" "}{link.icon}{" "}
                    </FontIcon>
                  }
                  onTouchTap={() => link.action(this.props.dispatch)}
                />
              );
            })}
          </IconMenu>
        </ToolbarGroup>
      </Toolbar>
    );
  }
}

export default connect(
  state => ({ search: getSearchState(state), menu: state.site.dropdownMenu }),
  (dispatch, props) => ({
    searchTextChanged: (txt, e) => dispatch(setSearchText(txt)),
    submit: () => dispatch(getSearchActionContext({ type: "SUBMIT" })),
    clearError: () =>
      dispatch(getSearchActionContext({ type: "SET_ERROR", value: "" })),
    dispatch: dispatch
  })
)(TopBar);
