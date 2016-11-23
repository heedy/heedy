export const UserEditInitialState = {};

export default function userEditReducer(state, action) {
    switch (action.type) {
        case 'USER_EDIT_CLEAR':
            return UserEditInitialState;
        case 'USER_EDIT':
            return Object.assign({},state,action.value);
    }
    return state;
}
