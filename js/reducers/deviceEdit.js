export const DeviceEditInitialState = {};

export default function deviceEditReducer(state, action) {
    switch (action.type) {
        case 'DEVICE_EDIT_CLEAR':
            return DeviceEditInitialState;
        case 'DEVICE_EDIT_NICKNAME':
            return {
                ...state,
                nickname: action.value
            };
        case 'DEVICE_EDIT_DESCRIPTION':
            return {
                ...state,
                description: action.value
            };
        case 'DEVICE_EDIT_PASSWORD':
            return {
                ...state,
                password: action.value
            };
        case 'DEVICE_EDIT_PASSWORD2':
            return {
                ...state,
                password2: action.value
            };
        case 'DEVICE_EDIT_ROLE':
            return {
                ...state,
                role: action.value
            };
        case 'DEVICE_EDIT_PUBLIC':
            return {
                ...state,
                public: action.value
            };
        case 'DEVICE_EDIT_APIKEY':
            // The API key can be set or reset
            let newval = {
                ...state
            };
            if (action.value) {
                newval.apikey = "";
            } else {
                delete newval.apikey;
            }
            return newval
        case 'DEVICE_EDIT_ENABLED':
            return {
                ...state,
                enabled: action.value
            };
    }
    return state;
}
