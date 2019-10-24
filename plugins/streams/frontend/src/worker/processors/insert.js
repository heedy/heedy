async function insert(source, data) {
    if (source == null || source.access != "*" || Object.keys(source.meta.schema).length == 0) {
        return {};
    }
    return {
        insert: {
            view: "insert",
            title: "Insert",
            data: {
                schema: source.meta.schema,
                id: source.id
            },
            weight: 0
        }
    }
}

export default insert;