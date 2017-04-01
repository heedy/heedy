import { StreamSearchInitialState, streamSearchReducer } from './search';

import moment from 'moment';
export const StreamViewInitialState = {
    expanded: false,
    transform: "",
    t1: moment().subtract(7, 'days'),
    t2: moment().endOf('day'),
    i1: -50,
    i2: 0,
    limit: 100000,
    data: [],
    error: null,
    bytime: true,
    views: {},
    loading: true,
    search: StreamSearchInitialState,
    firstvisit: true // We want to auto-load stream data on first visit only.
};

export default function streamViewReducer(state, action) {
    if (action.type.startsWith("STREAM_VIEW_SEARCH_"))
        return {
            ...state,
            search: streamSearchReducer(state.search, action)
        };

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
                error: null,
                loading: false
            };
        case 'STREAM_VIEW_ERROR':
            return {
                ...state,
                error: action.value,
                loading: false
            };
        case 'STREAM_VIEW_LOADING':
            return {
                ...state,
                loading: action.value
            };
    }
    return state;
}
