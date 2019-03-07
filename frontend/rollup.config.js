import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import postcss from "rollup-plugin-postcss";
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
  resolve(),
  postcss(),
  replace({
    "process.env.NODE_ENV": JSON.stringify(production ? "production" : "debug")
  })
];
if (production) {
  plugins.push(terser());
}
function externalize(arr) {
  // Add all generated outputs as valid externals
  let externals = Object.keys(globals);
  arr.map(o => {
    externals.push(
      "./" +
        o.output.file.substring(
          "../assets/public/".length,
          o.output.file.length
        )
    );
  });

  arr.map(o => {
    o.external = externals;
  });

  return arr;
}

function out(name, loc = "heedy/", format = "es") {
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
    plugins: plugins
    //external: ["vue", "./frontend/theme.mjs"]
  };
}

export default externalize([
  // The base files
  out("main.js", ""),
  out("404.vue"),
  out("loading.vue"),
  out("theme.vue"),
  out("login.vue"),
  // Default components
  out("user.vue")
]);
