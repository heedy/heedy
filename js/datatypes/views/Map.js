/*
Map shows a map with
*/

import { addView } from '../datatypes';

import React, { Component, PropTypes } from 'react';

import DataTransformUpdater from './components/DataUpdater';

import L from 'leaflet';
import { Map, Circle, Popup, TileLayer } from 'react-leaflet';

// The loader fails on leaflet css, so it is included manually in the template
//import 'leaflet/dist/leaflet.css';

import moment from 'moment';

import { location, objectvalues } from './typecheck';

class MapViewComponent extends DataTransformUpdater {

    preprocessData(d) {
        // We want to allow displaying numeric data on the map in the form of color. To do this, our datapoints
        // might be in a format that is NOT latitude/longitude, but {key1,key2}, where one of the keys is a number,
        // and the other key is the location.
        // We also want to validate all the datapoints to make sure they can be used for the map.

        let dataset = new Array(d.length);

        // First, check if it is lat/long or not. We have functions that map the datapoint correctly, with a "color"
        // option that allows us to set a magnitude
        let latlong = (d) => d;
        let color = (d) => 0;
        if (location(d) === null) {
            // We need to find which key is latitude and longitude
            let v = objectvalues(d);
            let keys = Object.keys(v);
            if (v[keys[0]].location !== null) {
                latlong = (d) => d[keys[0]];
                color = (d) => d[keys[1]];
            } else {
                latlong = (d) => d[keys[1]];
                color = (d) => d[keys[0]];
            }
        }

        // Now generate the dataset
        let j = 0;
        let minColor = 9999999999999;
        let maxColor = -minColor;
        for (let i = 0; i < d.length; i++) {
            let dp = latlong(d[i].d);
            if (dp.accuracy !== undefined && dp.accuracy > 50) {
                // Ignore the datapoint, since it is either invalid or inaccurate
                //console.log("ignoring", dp);
            } else {
                let c = color(d[i].d);
                if (c < minColor) minColor = c;
                if (c > maxColor) maxColor = c;
                dataset[j] = {
                    latlong: [dp.latitude, dp.longitude],
                    radius: (dp.accuracy !== undefined ? dp.accuracy : 20),
                    magnitude: c,
                    t: d[i].t
                }
                j += 1;
            }
        }

        dataset = dataset.slice(0, j);

        // Now, normalize the magnitudes of color to the range of 0 to 1
        if (minColor == maxColor) {
            for (let i = 0; i < dataset.length; i++) {
                dataset[i].magnitude -= minColor;
            }
        } else {
            for (let i = 0; i < dataset.length; i++) {
                dataset[i].magnitude = (dataset[i].magnitude - minColor) / (maxColor - minColor);
            }
        }
        return dataset;
    }
    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        // We can't graph all the datapoints - when there are more than ~1000 datapoints Leaflet slows to
        // a total crawl. We will therefore sift through the datapoints so that we are showing UP TO the max number
        // of datapoints at all times.
        let maxDatapoints = 1000;
        let d2 = this.preprocessData(d);
        let dataset = new Array(d2.length);

        let fillopacity = (d2.length > 200 ? 0.1 : d2.length > 50 ? 0.3 : 0.5);
        let opacity = (d2.length > 300 ? 0.2 : d2.length > 50 ? 0.5 : 0.9);
        for (let i = 0; i < d2.length; i++) {
            dataset[i] = {
                ...d2[i],
                color: `hsl(${Math.floor(120 * d2[i].magnitude)},100%,50%)`,
                fillopacity: fillopacity,
                opacity: opacity,
                key: JSON.stringify(d2[i]),
                popup: moment.unix(d2[i].t).calendar() + " - [" + d2[i].latlong[0].toString() + "," + d2[i].latlong[1].toString() + "]",
            }
        }

        return dataset;
    }
    render() {

        return (
            <Map center={this.data[this.data.length - 1].latlong} zoom={14} style={{
                width: '100%',
                height: '600px'
            }}>
                <TileLayer key="tileLayer" url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png' attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors' />
                {this.data.map((d) => (
                    <Circle key={d.key} center={d.latlong} radius={d.radius} color={d.color} fillOpacity={d.fillopacity} opacity={d.opacity} weight={2}>
                        <Popup>
                            <p>{d.popup}</p>
                        </Popup>
                    </Circle>
                ))}
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
        if (location(context.data) !== null) {
            return MapView;
        }

        // There is another option. if there are only 2 keys, and one is a location,
        // we can display a map with color-coded magnitude of the second key located
        // on the map. This is especially useful for datasets where one of 2 streams
        // is location!
        let v = objectvalues(context.data);
        if (v !== null) {
            let keys = Object.keys(v);

            if (keys.length == 2) {
                if (v[keys[0]].location && v[keys[1]].numeric) {
                    return MapView;
                } else if (v[keys[1]].location && v[keys[0]].numeric) {
                    return MapView;
                }
            }
        }

    }

    return null;
}

addView(showMap);
