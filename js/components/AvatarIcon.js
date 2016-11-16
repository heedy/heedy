import React, {Component, PropTypes} from 'react';
import Avatar from 'material-ui/Avatar';
import FontIcon from 'material-ui/FontIcon';

// The following code is used to generate icon colors for users/devices/streams
// https://stackoverflow.com/questions/3426404/create-a-hexadecimal-colour-based-on-a-string-with-javascript
function hashString(str) { // java String#hashCode
    var hash = 0;
    for (var i = 0; i < str.length; i++) {
        hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    return hash;
}

function intToRGB(i) {
    var c = (i & 0x00FFFFFF).toString(16).toUpperCase();

    return "#00000".substring(0, 7 - c.length) + c;
}

function stringToColor(str) {
    return intToRGB(hashString(str + str + str));
}

class AvatarIcon extends Component {
    static propTypes = {
        name: PropTypes.string.isRequired,
        iconsrc: PropTypes.string
    }
    render() {
        const {
            iconsrc,
            name,
            ...rest
        } = this.props;
        if (iconsrc !== undefined && iconsrc != "") {
            // Show the icon image if it exists

            //If the image starts with material: it means that we want to show the material icon
            if (iconsrc.startsWith("material:")) {
                return (<Avatar {...rest} backgroundColor={stringToColor(name)}  icon={<FontIcon className="material-icons">{iconsrc.substring("material:".length,iconsrc.length)}</FontIcon>} />);
            }

            // Otherwise, we use assume the image is URL encoded
            return (<Avatar {...rest} backgroundColor={stringToColor(name)} src={iconsrc} />);
        }
        return (
            <Avatar {...rest} backgroundColor={stringToColor(name)}>{name.substring(0, 1).toUpperCase()}</Avatar>
        );
    }
}

export default AvatarIcon;
