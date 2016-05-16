const InitialState = {
    queryText: ""
};

export default function queryReducer(state = InitialState, action) {
    switch (action.type) {
        case 'SET_QUERY_TEXT':
            return {
                ...state,
                queryText: action.value
            };
    }
    return state;
}
