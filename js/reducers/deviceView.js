export const DeviceViewInitialState = {
    expanded: true
};

export default function deviceViewReducer(state, action) {
    switch (action.type) {
        case 'DEVICE_VIEW_EXPANDED':
            return {
                ...state,
                expanded: action.value
            }
    }
    return state;
}
