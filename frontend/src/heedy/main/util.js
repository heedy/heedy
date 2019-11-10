function filterComponents(components, importance, filter) {
    let cmps = Object.values(components.filter(filter).reduce((o, v) => {
        // Multiple candidates might have the same key. We can't just choose the latest one,
        // though, since an older key might actually be more specific (for example, a source header
        // for a specific plugin/key combo should not be replaced by a source type header, even if it comes later)
        let newv = {
            importance: Object.keys(v).reduce(
                (w, k) => w + (importance[k] || 0),
                0
            ),
            ...v
        };
        if (
            o[v.key] === undefined ||
            o[v.key].importance <= newv.importance
        ) {
            if (newv.weight === undefined) {
                newv.weight = o[v.key].weight;
            }
            o[v.key] = newv;
        }
        return o;
    }, {})).filter((c) => c.component !== undefined && c.component !== null);
    cmps.sort((a, b) => a.weight - b.weight);
    return cmps;
}

export {
    filterComponents
};