import moment from 'moment';
export const StreamViewInitialState = {
    expanded: false,
    transform: "",
    t1: moment().subtract(7, 'days'),
    t2: moment().endOf('day'),
    i1: -50,
    i2: 0,
    limit: 0,
    data: [],
    error: null,
    bytime: true,
    views: {}
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
