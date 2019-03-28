import glob from "glob";
import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import postcss from "rollup-plugin-postcss";
import json from "rollup-plugin-json";
import externalGlobals from "rollup-plugin-external-globals";
import VuePlugin from "rollup-plugin-vue";
import replace from "rollup-plugin-replace";
import { terser } from "rollup-plugin-terser";

let globals = {
  vue: "Vue"
};

const production = !process.env.ROLLUP_WATCH;
const plugins = [
  VuePlugin(),
  commonjs(),
  // globals are not handled correctly by rollup, usually needing shim modules, which is BS
  //
  // https://github.com/rollup/rollup/issues/1437
  // https://github.com/rollup/rollup/issues/2374
  externalGlobals(globals),
  resolve({
    browser: true,
    preferBuiltins: false
  }),
  postcss(),
  json({
    compact: true
  }),
  replace({
    "process.env.NODE_ENV": JSON.stringify(production ? "production" : "debug")
  })
];
if (production) {
  plugins.push(terser());
}

function checkExternal(modid, parent, isResolved) {
  return (!isResolved && modid.endsWith(".mjs")) || modid.startsWith("http");
}
/*
function externalize(arr) {
  // Add all generated outputs as valid externals
  let externals = Object.keys(globals);
  arr.map(o => {
    externals.push(
      "./" +
        o.output.file.substring(
          "../assets/public/static/".length,
          o.output.file.length
        )
    );
  });

  arr.map(o => {
    o.external = externals;
  });
  console.log(arr);
  return arr;
}
*/

function out(name, loc = "", format = "es") {
  let filename = name.split(".");
  return {
    input: "src/" + name,
    output: {
      name: filename[0],
      file:
        "../assets/public/static/" +
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
  out("auth.js"),
  out("setup.js"),
  out("heedy/util.js")
]
  .concat(glob.sync("heedy/*.vue", { cwd: "./src" }).map(a => out(a)))
  .concat(
    glob.sync("heedy/components/*.vue", { cwd: "./src" }).map(a => out(a))
  );
