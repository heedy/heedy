import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import postcss from "rollup-plugin-postcss";
import externalGlobals from "rollup-plugin-external-globals";
import VuePlugin from "rollup-plugin-vue";
import replace from "rollup-plugin-replace";
import { terser } from "rollup-plugin-terser";

const production = !process.env.ROLLUP_WATCH;
const plugins = [
  VuePlugin(),
  commonjs(),
  externalGlobals({
    vue: "Vue"
  }),
  resolve(),

  postcss(),
  // globals are not handled correctly by rollup, usually needing shim modules, which is BS
  //
  // https://github.com/rollup/rollup/issues/1437
  // https://github.com/rollup/rollup/issues/2374

  replace({
    "process.env.NODE_ENV": JSON.stringify(production ? "production" : "debug")
  })
];
if (production) {
  plugins.push(terser());
}
function out(name, loc = "js/", format = "es") {
  let filename = name.split(".");
  return {
    input: "src/" + name,
    output: {
      name: filename[0],
      file:
        "../assets/public/" +
        loc +
        filename[0] +
        (format == "es" ? ".mjs" : ".js"),
      format: format
    },
    plugins: plugins,
    external: ["vue", "./js/theme.mjs"]
  };
}

export default [
  // The base files
  out("main.js", ""),
  out("theme.vue"),
  // Default components
  out("user.vue")
];
