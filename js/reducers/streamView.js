export const StreamViewInitialState = {
    expanded: false,
    tExpanded: false,
    fullwidth: false,
    transform: "",
    last: 10,
    t1: 0,
    t2: 0,
    limit: 10,
    data: [],
    error: null,
    bytime: false
};

export default function streamViewReducer(state, action) {
    switch (action.type) {
        case 'STREAM_VIEW_EXPANDED':
            return {
                ...state,
                expanded: action.value
            };
        case 'STREAM_VIEW_SET':
            return Object.assign({}, state, action.value);
        case 'STREAM_VIEW_DATA':
            return {
                ...state,
                data: action.value,
                error: null
            };
        case 'STREAM_VIEW_ERROR':
            return {
                ...state,
                error: action.value
            };
    }
    return state;
}
