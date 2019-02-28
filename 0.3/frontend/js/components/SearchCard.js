/*
  This component is colored like the search bar to make it clear that stuff inside it is related to search
*/

import React, { Component } from "react";
import PropTypes from "prop-types";
import { Card, CardText, CardHeader } from "material-ui/Card";
import FontIcon from "material-ui/FontIcon";
import IconButton from "material-ui/IconButton";

class SearchCard extends Component {
  static propTypes = {
    title: PropTypes.string,
    subtitle: PropTypes.string,
    color: PropTypes.string,
    onClose: PropTypes.func
  };
  static defaultProps = {
    color: "#00b34a"
  };

  render() {
    if (this.props.onClose !== undefined) {
    }
    return (
      <Card
        style={{
          marginTop: "25px",
          textAlign: "left",
          marginBottom: "20px",
          backgroundColor: this.props.color
        }}
      >
        {this.props.title !== undefined
          ? <CardHeader
              title={this.props.title}
              titleColor="white"
              titleStyle={{
                fontWeight: "bold"
              }}
              subtitle={this.props.subtitle}
            >
              {this.props.onClose !== undefined
                ? <div
                    style={{
                      float: "right",
                      marginRight: "0px",
                      marginTop: this.props.subtitle === undefined
                        ? "-15px"
                        : "-5px",
                      marginLeft: "-300px"
                    }}
                  >
                    <IconButton onTouchTap={() => this.props.onClose()}>
                      <FontIcon
                        className="material-icons"
                        color="rgba(0,0,0,0.5)"
                      >
                        close
                      </FontIcon>
                    </IconButton>
                  </div>
                : null}

            </CardHeader>
          : null}
        {" "}
        {this.props.children !== undefined && this.props.children.length > 0
          ? <CardText
              style={{
                textAlign: "center",
                color: "white"
              }}
            >
              {this.props.children}
            </CardText>
          : null}
      </Card>
    );
  }
}
export default SearchCard;
