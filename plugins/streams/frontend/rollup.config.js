
import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import postcss from "rollup-plugin-postcss";
import json from "rollup-plugin-json";
import VuePlugin from "rollup-plugin-vue";
import replace from "rollup-plugin-replace";
import { terser } from "rollup-plugin-terser";

const production = !process.env.NODE_ENV==='debug';
const plugins = [
  VuePlugin(),
  commonjs(),
  resolve({
    browser: true,
    preferBuiltins: false
  }),
  postcss({
    minimize: production
  }),
  json({
    compact: production
  }),
  replace({
    "process.env.NODE_ENV": JSON.stringify(production ? "production" : "debug")
  })
];
if (production) {
  plugins.push(terser({
    compress:{
      drop_console: true,
      ecma: 10 // Heedy doesn't do backwards compatibility
    },
    mangle: true,
    module: true
  }));
}

function checkExternal(modid, parent, isResolved) {
  return (!isResolved && modid.endsWith(".mjs")) || modid.startsWith("http");
}

function out(name, loc = "", format = "es") {
  let filename = name.split(".");
  return {
    input: "src/" + name,
    output: {
      name: filename[0],
      file:
        "../assets/public/static/streams/" +
        loc +
        filename[0] +
        (format == "es" ? ".mjs" : ".js"),
      format: format
    },
    plugins: plugins,
    external: checkExternal
  };
}
export default [
  // The base files
  out("main.js"),
  out("preprocessing.worker.js")
];
