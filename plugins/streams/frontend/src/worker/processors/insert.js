async function insert(source, data) {
    if (source == null || source.access != "*") {
        return {};
    }
    return {
        insert: {
            component: "insert",
            data: {},
            weight: 0
        }
    }
}

export default insert;