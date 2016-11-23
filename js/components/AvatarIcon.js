import React, {Component, PropTypes} from 'react';
import Avatar from 'material-ui/Avatar';
import FontIcon from 'material-ui/FontIcon';

// The following code is used to generate icon colors for users/devices/streams
// Originally, comors were generated completely at random, but that was not perfect.
// See: http://martin.ankerl.com/2009/12/09/how-to-create-random-colors-programmatically/
// Fortunately, HSL is supported directly in browsers these days.

//http://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript-jquery
function hashFnv32a(str) {
    /*jshint bitwise:false */
    var i, l;
    let hval =  0x811c9dc5;

    for (i = 0, l = str.length; i < l; i++) {
        hval ^= str.charCodeAt(i);
        hval += (hval << 1) + (hval << 4) + (hval << 7) + (hval << 8) + (hval << 24);
    }
    return hval >>> 0;
}

function intToHSL(i) {
    i = i % 360;
    return "hsl("+i.toString()+",53%,45%)";
}

function stringToColor(str) {
    return intToHSL(hashFnv32a(str));
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
                return (<Avatar {...rest} backgroundColor={stringToColor(iconsrc+name)}  icon={<FontIcon className="material-icons">{iconsrc.substring("material:".length,iconsrc.length)}</FontIcon>} />);
            }

            if (iconsrc.startsWith("data:image/")) {
                // We use assume the image is URL encoded
                return (<Avatar {...rest} backgroundColor={stringToColor(iconsrc+name)} src={iconsrc} />);
            }
            
        }
        return (
            <Avatar {...rest} backgroundColor={stringToColor(name)}>{name.substring(0, 1).toUpperCase()}</Avatar>
        );
    }
}

export default AvatarIcon;
