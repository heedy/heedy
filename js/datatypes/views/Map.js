/*
Map shows a map with
*/

import {addView} from '../datatypes';

import React, {Component, PropTypes} from 'react';

import DataTransformUpdater from './components/DataUpdater';

import L from 'leaflet';
import {Map, Marker, Popup, TileLayer} from 'react-leaflet';

// The loader fails on leaflet css, so it is included manually in the template
//import 'leaflet/dist/leaflet.css';

import moment from 'moment';

// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

// Checks if the given datapoint is latitude/longitude
function isLatLong(point) {
    return (point.latitude !== undefined && isNumeric(point.latitude) && point.longitude !== undefined && isNumeric(point.longitude));
}

var mapdot = L.icon({
    iconUrl: SiteURL + '/app/css/red_dot.png',
    iconSize: [
        10, 10
    ],
    iconAnchor: [
        5, 5
    ],
    popupAnchor: [0, 0]
});

class MapViewComponent extends DataTransformUpdater {
    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        // We can't graph all the datapoints - when there are more than ~1000 datapoints Leaflet slows to
        // a total crawl. We will therefore sift through the datapoints so that we are showing UP TO 1000 datapoints
        // at all times
        let maxDatapoints = 1000;
        let datapointChange = maxDatapoints / d.length;

        let dataset = new Array(d.length > maxDatapoints
            ? maxDatapoints
            : d.length);
        let j = 0;
        let prevdpnum = -2;
        let dpnum = 0;
        for (let i = 0; i < d.length; i++) {
            dpnum += datapointChange;
            if (Math.floor(prevdpnum) + 1 <= Math.floor(dpnum)) {
                // We process the datapoint
                prevdpnum = dpnum;
                if (!isLatLong(d[i].d) || d[i].accuracy !== undefined && d[i].accuracy > 100) {
                    // We just ignore this one, since it is inaccurate or not valid
                } else {
                    dataset[j] = {
                        ...d[i].d,
                        popup: moment.unix(d[i].t).calendar() + " - [" + d[i].d.latitude.toString() + "," + d[i].d.longitude.toString() + "]",
                        key: JSON.stringify(d[i])
                    };
                    j++;
                }
            }
        }
        return dataset.slice(0, j);
    }
    render() {
        var markers = [(<TileLayer key="tileLayer" url='http://{s}.tile.osm.org/{z}/{x}/{y}.png' attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'/>)].concat(this.data.map((d) => (
            <Marker key={d.key} icon={mapdot} position={[d.latitude, d.longitude]} opacity={0.7}>
                <Popup>
                    <p>{d.popup}</p>
                </Popup>
            </Marker>
        )));

        return (
            <Map center={[
                this.data[this.data.length - 1].latitude,
                this.data[this.data.length - 1].longitude
            ]} zoom={13} style={{
                width: '100%',
                height: '600px'
            }}>
                {markers}

            </Map>
        );
    }
}

const MapView = {
    key: "mapView",
    component: MapViewComponent,
    width: "expandable-full",
    initialState: {},
    title: "Map",
    subtitle: ""
}

function showMap(context) {
    if (context.data.length > 0) {
        // We now check if the data is the correct type
        if (isLatLong(context.data[0].d) && isLatLong(context.data[context.data.length - 1].d)) {
            return MapView;
        }

    }

    return null;
}

addView(showMap);
