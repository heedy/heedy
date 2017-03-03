var webpack = require('webpack');
var path = require('path');

if (Promise === undefined) {
    require('es6-promise').polyfill();
}

// Use the ConnectorDB bin directory to output files for easy debugging
var BUILD_DIR = path.resolve(__dirname, '../../bin/app');
var APP_DIR = path.resolve(__dirname, 'js');
var env = process.env.NODE_ENV
var config = {
    entry: APP_DIR + '/index.js',
    output: {
        path: BUILD_DIR,
        filename: 'bundle.js',
        libraryTarget: "var",
        library: "App",
        publicPath: "/app/"
    },

    module: {
        //noParse: [path.join(__dirname, "node_modules", "pipescript")],
        rules: [
            {
                test: /\.jsx?/,
                include: APP_DIR,
                loader: 'babel-loader'
            }, {
                test: /\.css$/,
                use: ["style-loader","css-loader"]
            }
        ]
    },

    plugins: [
        new webpack.DefinePlugin({'process.env.NODE_ENV': JSON.stringify(env)})
    ]
};

if (env === 'production') {
    config.plugins.push(new webpack.optimize.DedupePlugin())
    config.plugins.push(new webpack.optimize.UglifyJsPlugin({
        compressor: {
            pure_getters: true,
            unsafe: true,
            unsafe_comps: true,
            screw_ie8: true,
            warnings: false
        }
    }))
}

module.exports = config;
