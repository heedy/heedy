export const UserViewInitialState = {
    expanded: true,
    hiddel: true
};

export default function userViewReducer(state, action) {
    switch (action.type) {
        case 'USER_VIEW_EXPANDED':
            return {
                ...state,
                expanded: action.value
            }
        case 'USER_VIEW_HIDDEN':
            return {
                ...state,
                hidden: action.value
            }
    }
    return state;
}
