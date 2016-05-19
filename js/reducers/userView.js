export const UserViewInitialState = {
    expanded: true
};

export default function userViewReducer(state, action) {
    switch (action.type) {
        case 'USER_VIEW_EXPANDED':
            return {
                ...state,
                expanded: action.value
            }
    }
    return state;
}
