/**
 * The HeatmapView shows various heatmaps thanks to reactochart
 */
import React, { Component, PropTypes } from 'react';

import { addView } from '../datatypes';
import { objectvalues, dataKeyCompare } from './typecheck';
import DataTransformUpdater from './components/DataUpdater';

import {XYPlot, XGrid, YGrid, XAxis, YAxis} from 'reactochart';
import ColorHeatmap from 'reactochart/lib/ColorHeatmap';
import Measure from 'react-measure';

console.log("heatmap",ColorHeatmap);
import _ from 'lodash';

function convMap(m) {
    let o = {};
    m.forEach((v,k) => o[String(k)] = v);
    return o;
}
// Based on:
// https://github.com/spotify/reactochart/blob/0.2.1/examples/src/Examples.jsx
class HeatmapViewComponent extends DataTransformUpdater {
    constructor(p) {
        super(p);
        this.state = {
            width: 300
        };
    }
// transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        // Both objects are categorical
        let o = objectvalues(d);

        // We want to first sort the categories correctly
        let k = Object.keys(o);
        this.k = k;

        let xobj = convMap(o[k[0]].categorical.categorymap);
        let yobj = convMap(o[k[1]].categorical.categorymap);

        this.xkeys = Object.keys(xobj).sort(dataKeyCompare.bind(xobj));
        this.ykeys = Object.keys(yobj).sort(dataKeyCompare.bind(yobj));

        // Now go through the data to find which datapoints fit into which category 
        let catmatcher = {};
        for (let i=0;i < d.length;i++) {
            let x = String(d[i].d[k[0]]);
            let y = String(d[i].d[k[1]]);

            if (catmatcher[x]===undefined) catmatcher[x] = {};
            if (catmatcher[x][y]===undefined) catmatcher[x][y] = 0;

            catmatcher[x][y]++;

        }
        

        return _.flatten(this.xkeys.map((n,i)=> this.ykeys.map((m,j)=> ({
            x:i,
            xEnd: i+1,
            y: j,
            yEnd: j+1,
            value: catmatcher[n][m]===undefined?0:catmatcher[n][m]
        }))));
    }

    render() {
        return (
            <div style={{paddingRight: 20}}>
        <Measure onMeasure={({width})=> this.setState({width:width})}>
            <div>
                <XYPlot width={this.state.width} height={this.state.width<800?this.state.width:800} >
                    <ColorHeatmap
                        data={this.data}
                        getValue="value"
                        getX="x"
                        getXEnd="xEnd"
                        getY="y"
                        getYEnd="yEnd"
                        colors={['rebeccapurple', 'goldenrod']}
                        interpolator={'lab'}
                    />
                    <XAxis
                        showGrid={false}
                        ticks={this.xkeys.map((t, i) => i + 0.5)}
                        title={this.k[0]}
                        labelFormats={this.xkeys}
                    />
                    <YAxis
                        showGrid={false}
                        ticks={this.ykeys.map((t, i) => i + 0.5)}
                        title={this.k[1]}
                        labelFormats={this.ykeys}
                    />
                    <XGrid tickCount={4} />
                    <YGrid tickCount={4} />
                </XYPlot>
            </div>
        </Measure>
        </div>
        );
    }

}


const HeatmapView = {
    key: "heatmapView",
    component: HeatmapViewComponent,
    width: "expandable-half",
    initialState: {},
    title: "Heatmap",
    subtitle: ""
}
  

function showHeatmap(context) {
    if (context.data.length < 20 || context.pipescript === null) {
        return null;
    }

    let o = objectvalues(context.data);
    if (o !== null && Object.keys(o).length == 2) {
        let k = Object.keys(o);

        // Make sure that the data is actually categorical
        if (o[k[0]].categorical !== null && o[k[1]].categorical !== null && o[k[0]].categorical.categories < 50 && o[k[1]].categorical.categories < 50) {

            // We can display the plot!
            return HeatmapView;
        }
    }
    return null;
}

addView(showHeatmap);
