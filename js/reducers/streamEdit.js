export const StreamEditInitialState = {};

export default function streamEditReducer(state, action) {
    switch (action.type) {
        case 'STREAM_EDIT_CLEAR':
            return StreamEditInitialState;
        case 'STREAM_EDIT_NICKNAME':
            return {
                ...state,
                nickname: action.value
            };
        case 'STREAM_EDIT_DESCRIPTION':
            return {
                ...state,
                description: action.value
            };
        case 'STREAM_EDIT_DOWNLINK':
            return {
                ...state,
                downlink: action.value
            };
        case 'STREAM_EDIT_EPHEMERAL':
            return {
                ...state,
                ephemeral: action.value
            };
        case 'STREAM_EDIT_DATATYPE':
            return {
                ...state,
                datatype: action.value
            };
    }
    return state;
}
