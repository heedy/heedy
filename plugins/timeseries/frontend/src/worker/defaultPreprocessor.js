import jjt from "../../dist/json-json-template.mjs";

function jjtAnalyzer(context,config) {
    let jf = jjt(config);
    return jf(context);
}

export default jjtAnalyzer;