const path = require("path");
module.exports = {
  outputDir: "../assets/setup",
  publicPath: "",
  //assetsDir: "static"

  // Don't add the hashes to files, so that plugins can
  // override them if necessary.
  //https://github.com/vuejs/vue-cli/issues/1649
  configureWebpack: {
    output: {
      filename: "[name].js",
      chunkFilename: "[name].js"
    }
  },
  chainWebpack: config => {
    if (config.plugins.has("extract-css")) {
      const extractCSSPlugin = config.plugin("extract-css");
      extractCSSPlugin &&
        extractCSSPlugin.tap(() => [
          {
            filename: "[name].css",
            chunkFilename: "[name].css"
          }
        ]);
    }

    config.externals(/p\/.*\.js/);
  }
};
