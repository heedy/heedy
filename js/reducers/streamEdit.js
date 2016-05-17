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
        case 'STREAM_EDIT_PASSWORD':
            return {
                ...state,
                password: action.value
            };
        case 'STREAM_EDIT_PASSWORD2':
            return {
                ...state,
                password2: action.value
            };
        case 'STREAM_EDIT_ROLE':
            return {
                ...state,
                role: action.value
            };
        case 'STREAM_EDIT_PUBLIC':
            return {
                ...state,
                public: action.value
            };
        case 'STREAM_EDIT_EMAIL':
            return {
                ...state,
                email: action.value
            };
    }
    return state;
}
