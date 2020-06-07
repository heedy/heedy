import resolve from "rollup-plugin-node-resolve"; // https://github.com/rollup/plugins/issues/67
import commonjs from "@rollup/plugin-commonjs";
import postcss from "rollup-plugin-postcss";
import json from "@rollup/plugin-json";
import VuePlugin from "rollup-plugin-vue";
import replace from "@rollup/plugin-replace";
import { terser } from "rollup-plugin-terser";
import gzipPlugin from "rollup-plugin-gzip";
//import { brotliCompressSync, constants } from "zlib";

import postcss_url from "postcss-url";
import path from "path";
import fs from "fs";
import glob from "glob";

let fontFolder = "../assets/public/static/fonts";
fs.mkdirSync(fontFolder, {
  recursive: true,
});

const production = !(process.env.NODE_ENV === "debug");
const plugins = [
  VuePlugin({
    // https://github.com/vuejs/rollup-plugin-vue/issues/238
    needMap: false,
  }),
  commonjs(),
  resolve({
    //browser: true,
    preferBuiltins: false,
  }),
  postcss({
    minimize: production,
    plugins: [
      postcss_url({
        // copy ALMOST does what we want - it renames the asset files... however,
        // what we ACTUALLY want is to move the files AND rename them relative to the root
        // so let's do that here
        url: function(asset, dir, options, decl, warn, result, addDependency) {
          if (asset.url.startsWith("data:")) {
            return asset.url;
          }
          // We really only need woff2 files, since we are targeting modern browsers
          if (path.extname(asset.absolutePath) == ".woff2") {
            let toURL = fontFolder + "/" + path.basename(asset.absolutePath);
            fs.copyFile(asset.absolutePath, toURL, (err) => {
              if (err) throw err;
              console.log(asset.relativePath, " -> ", toURL);
            });
          }

          return "/static/fonts/" + path.basename(asset.url);
        },
      }),
    ],
  }),
  json({
    compact: production,
  }),
  replace({
    "process.env.NODE_ENV": JSON.stringify(production ? "production" : "debug"),
  }),
];
if (production) {
  plugins.push(
    terser({
      compress: {
        drop_console: true,
        ecma: 10, // Heedy doesn't do backwards compatibility
      },
      mangle: true,
      module: true,
    })
  );
  /* WTF: firefox doesn't support brotli on localhost without https!!! We have to use gzip :(
  plugins.push(
    gzipPlugin({
      customCompression: (content) =>
        brotliCompressSync(Buffer.from(content), {
          params: {
            [constants.BROTLI_PARAM_MODE]: constants.BROTLI_MODE_TEXT,
            [constants.BROTLI_PARAM_QUALITY]: constants.BROTLI_MAX_QUALITY,
            [constants.BROTLI_PARAM_SIZE_HINT]: content.length,
          },
        }),
      fileName: ".br",
    })
  );
  */
  plugins.push(gzipPlugin());
} else {
  console.log("Running debug build");
}

function checkExternal(modid, parent, isResolved) {
  return (
    (!isResolved && modid.endsWith(".mjs") && modid.startsWith(".")) ||
    modid.startsWith("http")
  );
}

function out(name, loc = "", format = "es") {
  let filename = name.slice(0, name.lastIndexOf("."));
  return {
    input: "src/" + name,
    output: {
      name: filename,
      file:
        "../assets/public/static/" +
        loc +
        filename +
        (format == "es" ? ".mjs" : ".js"),
      format: format,
    },
    plugins: plugins,
    external: checkExternal,
  };
}

// g allows using globs to define output files
let g = (x) =>
  glob
    .sync(x, {
      cwd: "./src",
    })
    .map((a) => out(a));

export default [...g("*.js"), ...g("dist/*.js"), ...g("heedy/*.js")];
