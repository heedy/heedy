export const StreamViewInitialState = {
    expanded: true,
    data: []
};

export default function streamViewReducer(state, action) {
    switch (action.type) {
        case 'STREAM_VIEW_EXPANDED':
            return {
                ...state,
                expanded: action.value
            }
    }
    return state;
}
