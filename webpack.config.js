var webpack = require('webpack');
var path = require('path');

// Use the ConnectorDB bin directory to output files for easy debugging
var BUILD_DIR = path.resolve(__dirname, '../../bin/app');
var APP_DIR = path.resolve(__dirname, 'js');
var env = process.env.NODE_ENV
var config = {
    entry: APP_DIR + '/index.jsx',
    output: {
        path: BUILD_DIR,
        filename: 'bundle.js',
        libraryTarget: "var",
        library: "App"
    },

    module: {
        loaders: [{
            test: /\.jsx?/,
            include: APP_DIR,
            loader: 'babel'
        }]
    },

    plugins: [
        new webpack.optimize.OccurenceOrderPlugin(),
        new webpack.DefinePlugin({
            'process.env.NODE_ENV': JSON.stringify(env)
        })
    ]
};

if (env === 'production') {
    config.plugins.push(
        new webpack.optimize.UglifyJsPlugin({
            compressor: {
                pure_getters: true,
                unsafe: true,
                unsafe_comps: true,
                screw_ie8: true,
                warnings: false
            }
        })
    )
}

module.exports = config;
