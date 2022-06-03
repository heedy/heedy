import {cachedJJT} from "../../../dist/json-json-template.mjs";

// A map of functions that given a context and configuration, return the
// data necessary to show the visualization type.
const customPreprocessors = new Map();

function preprocessConfiguration(ctx,type,config) {
    if (customPreprocessors.has(type)) {
        return customPreprocessors.get(type)(ctx,type,config);
    }

    // By default, assume that the config is a json-json-template object,
    // so we can compile it. However, we generally want to avoid re-compiling
    // the same template, so we cache the compiled object by its json.
    const jf = cachedJJT(config);
    return jf(ctx);
}

// preprocess assumes that v has a key property
function preprocess(ctx, v) {
    try {
        return {
            ...v,
            data: preprocessConfiguration(ctx, v.type, v.config),
        };
    } catch (e) {
        return {
            type: "error",
            key: v.key,
            title: v.title!==undefined?v.title:v.key,
            config: v.config,
            weight: v.weight!==undefined?v.weight:1000,
            data: {
                title: "Error in Visualization Configuration",
                error: e,
                showConfig: true
            }
        };
    }
}

const preprocessAll = (ctx, vis) => Object.entries(vis).map(([key, v]) => preprocess(ctx, {...v, key}));

export {customPreprocessors, preprocessAll}
export default preprocess;