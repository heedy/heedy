async function insert(object, data) {
    if (object == null || object.access != "*" || Object.keys(object.meta.schema).length == 0) {
        return {};
    }
    return {
        insert: {
            view: "insert",
            title: "Insert",
            data: {
                schema: object.meta.schema,
                id: object.id
            },
            weight: 0
        }
    }
}

export default insert;