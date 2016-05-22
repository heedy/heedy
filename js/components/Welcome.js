// MainToolbar is the toolbar shown on the main page, from which you can create new rating streams,
// and so forth
import React, {Component, PropTypes} from 'react';

import {Card, CardText, CardHeader} from 'material-ui/Card';
import FontIcon from 'material-ui/FontIcon';
import IconButton from 'material-ui/IconButton';

import storage from '../storage';

class Welcome extends Component {
    static propTypes = {}

    render() {
        return (
            <Card style={{
                marginTop: "20px"
            }}>
                <CardHeader title={"Welcome!"} subtitle={"Get started with ConnectorDB"}/>
                <CardText>
                    <p>It looks like you don't have any ratings or manual inputs set up yet. Click on the star icon above (after reading this) to create your first rating stream.</p>
                    <p>Most data will be gathered automatically by your devices, such as an android app, or the laptop logger. Make sure to download the
                        <a href="https://connectordb.github.io/clients.html">{" "}open-source ConnectorDB clients{" "}</a>
                        if you have not done so already. You can also create your own devices using the
                        <a href="https://connectordb-python.readthedocs.io/en/latest/">{" "}python client{" "}</a>. Furthermore, you can control things, such as lights or your thermostat by creating and subscribing to downlink streams. If you ever get lost, please refer to the
                        <a href="https://connectordb.github.io/docs/">{" "}documentation{" "}</a>.
                    </p>
                    <p>Despite the automatic data gathering, our experience shows that it is very difficult to gain insight without having a supervision signal. This is where your manual input is useful. By creating ratings, you can rate your productivity, mood, or log other things, such as weight and progress towards goals. By consistently doing this every day, over time, you will gain enough data to be able to perform in-depth analysis on what exactly influences your abilities.
                    </p>

                </CardText>
            </Card>
        );
    }
}
export default Welcome;
