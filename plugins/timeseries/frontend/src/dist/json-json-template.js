import jjt from "json-json-template";
// Add a stable stringify for use in caching jjt outputs.
import stringify from "fast-json-stable-stringify";
const stableStringify = stringify;

const jjtCache = new Map();

function cachedJJT(template) {
    const hash = stableStringify(template);
    let outf = jjtCache.get(hash);
    if (outf !== undefined) {
        return outf;
    }
    outf = jjt(template);
    jjtCache.set(hash, outf);
    return outf;
}


export {stableStringify,cachedJJT};
export default jjt;
