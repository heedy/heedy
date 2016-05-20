export const StreamCreateInitialState = {
    name: "",
    nickname: "",
    description: "",
    schema: "",
    downlink: false,
    ephemeral: false
};

export default function deviceCreateReducer(state, action) {
    switch (action.type) {
        case 'DEVICE_CREATESTREAM_CLEAR':
            return StreamCreateInitialState;
        case 'DEVICE_CREATESTREAM_NAME':
            return {
                ...state,
                name: action.value
            };
        case 'DEVICE_CREATESTREAM_NICKNAME':
            return {
                ...state,
                nickname: action.value
            };
        case 'DEVICE_CREATESTREAM_DESCRIPTION':
            return {
                ...state,
                description: action.value
            };

    }
    return state;
}
