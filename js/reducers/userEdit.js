export const UserEditInitialState = {};

export default function userEditReducer(state, action) {
    switch (action.type) {
        case 'USER_EDIT_CLEAR':
            return UserEditInitialState;
        case 'USER_EDIT_NICKNAME':
            return {
                ...state,
                nickname: action.value
            };
        case 'USER_EDIT_DESCRIPTION':
            return {
                ...state,
                description: action.value
            };
        case 'USER_EDIT_PASSWORD':
            return {
                ...state,
                password: action.value
            };
        case 'USER_EDIT_PASSWORD2':
            return {
                ...state,
                password2: action.value
            };
        case 'USER_EDIT_ROLE':
            return {
                ...state,
                role: action.value
            };
        case 'USER_EDIT_PUBLIC':
            return {
                ...state,
                public: action.value
            };
        case 'USER_EDIT_EMAIL':
            return {
                ...state,
                email: action.value
            };
    }
    return state;
}
