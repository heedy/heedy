/*
An expandable card is used for inputs and views. It allows the card to be either half-width or full width based on
its width parameter and its state. The card also is set up for easy use with toolbar icons

For its state, it is given either undefined, or a specific value, in which case the state overrides size.
*/

import React, { Component, PropTypes } from 'react';

import { Card, CardText, CardHeader } from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';


class ExpandableCard extends Component {
    static propTypes = {
        // The size to show the card. One of:
        // half,full,expandable-half,expandable-full
        width: PropTypes.string.isRequired,

        // state: if given, overrides default
        // view info for the card
        state: PropTypes.object.isRequired,
        // Called to set the state
        setState: PropTypes.func.isRequired,

        // The dropdown is optional - if set, it shows the down arrow
        // allowing the component to show expanded options
        dropdown: PropTypes.element,

        // An optional array of icons to display
        icons: PropTypes.arrayOf(PropTypes.element),

        title: PropTypes.string.isRequired,
        subtitle: PropTypes.string.isRequired,
        style: PropTypes.object,

        avatar: PropTypes.element
    }

    render() {

        let state = this.props.state;
        let setState = this.props.setState;

        let width = this.props.width;

        let expandable = width.startsWith("expandable");
        if (expandable) {
            switch (width) {
                case "expandable":
                    width = "half";
                    break;
                case "expandable-half":
                    width = "half";
                    break;
                case "expandable-full":
                    width = "full";
                    break;
            }

            // If the card is expandable, we might have an overridden value in the state
            if (state.width !== undefined) {
                width = state.width;
            }
        }

        let hasDropdown = (this.props.dropdown !== undefined && this.props.dropdown !== null);

        // Now, we construct the icons to show on the right side of the card.
        // We have the dropdown icon (which is automatically shown if dropdown is activated)
        // we have the expand icon if the card can be switched between full and half width,
        // and finally, we have the optional array of icons that might have been passed in.
        // We therefore construct a large array of icons from the one we have

        let iconRightMargin = (hasDropdown
            ? 35
            : 0);
        let iconarray = [];
        if (this.props.icons !== undefined && this.props.icons != null) {
            iconarray = this.props.icons.slice(0);
        }
        // Now, if this card is expandable, add the expand icon/contract icon depending on which one is correct
        if (expandable) {
            if (width === "full") {
                iconarray.push((
                    <IconButton key="expand" onTouchTap={(val) => setState({ ...state, width: "half" })} tooltip="Make This Card Smaller" >
                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                            call_received
                        </FontIcon>
                    </IconButton >
                ));
            } else {
                iconarray.push((
                    <IconButton key="expand" onTouchTap={(val) => setState({ ...state, width: "full" })} tooltip="Expand to Full Width" >
                        <FontIcon className="material-icons" color="rgba(0,0,0,0.8)">
                            call_made
                        </FontIcon >
                    </IconButton >
                ))
            }
        }

        return (
            <div className={width === "full"
                ? "col-lg-12"
                : "col-lg-6"}>
                <Card style={{
                    marginTop: "20px",
                    textAlign: "left"
                }} onExpandChange={(val) => setState({
                    ...state,
                    expanded: val
                })} expanded={state.expanded}>
                    <CardHeader title={this.props.title} subtitle={this.props.subtitle} showExpandableButton={hasDropdown} avatar={this.props.avatar}>
                        <div style={{
                            float: "right",
                            marginRight: iconRightMargin,
                            marginTop: (this.props.subtitle == ""
                                ? "-15px"
                                : "-3px"),
                            marginLeft: "-300px"
                        }}>
                            {iconarray}
                        </div>
                    </CardHeader>
                    {hasDropdown
                        ? (
                            <CardText expandable={true} style={{
                                backgroundColor: "rgba(0,179,74,0.05)",
                                paddingBottom: "30px"
                            }}>
                                {this.props.dropdown}
                            </CardText>
                        )
                        : null}

                    <CardText style={this.props.style}>
                        {this.props.children}
                    </CardText>
                </Card>
            </div >
        );

    }

}

export default ExpandableCard;
