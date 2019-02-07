import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import VuePlugin from "rollup-plugin-vue";
import { terser } from "rollup-plugin-terser";

function out(name, format = "es") {
  let filename = name.split(".");
  return {
    input: "src/" + name,
    output: {
      name: filename[0],
      file:
        "../assets/app/js/" + filename[0] + (format == "es" ? ".jsm" : ".js"),
      format: format
    },
    plugins: [resolve(), commonjs(), VuePlugin(), terser()]
  };
}

export default [
  // The base files
  out("base.js", "umd"),
  out("theme.js", "umd"),
  // Default components
  out("user.vue")
];
