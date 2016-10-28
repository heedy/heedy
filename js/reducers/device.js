// The deviceReducer maintains state for all devices - each page that was visited in this session has
// its state maintained here. Each action that operates on a device contains a "name" field, which states the device
// and device in whose context to operate

// The keys in this object are going to be the device paths
const InitialState = {};

import deviceViewReducer, {DeviceViewInitialState} from './deviceView';
import deviceEditReducer, {DeviceEditInitialState} from './deviceEdit';
import streamCreateReducer, {StreamCreateInitialState} from './streamCreate';

// The initial state of a specific device
const DeviceInitialState = {
    edit: DeviceEditInitialState,
    view: DeviceViewInitialState,
    create: StreamCreateInitialState
};

export default function deviceReducer(state = InitialState, action) {
    if (!action.type.startsWith("DEVICE_"))
        return state;

    // Set up the new state
    let newState = {
        ...state
    };

    // If the device already has a state, copy the next level, otherwise, initialize the next level
    if (state[action.name] !== undefined) {
        newState[action.name] = {
            ...state[action.name]
        }
    } else {
        newState[action.name] = Object.assign({}, DeviceInitialState);
    }

    // Now route to the appropriate reducer
    if (action.type.startsWith("DEVICE_EDIT_"))
        newState[action.name].edit = deviceEditReducer(newState[action.name].edit, action);
    if (action.type.startsWith("DEVICE_VIEW_"))
        newState[action.name].view = deviceViewReducer(newState[action.name].view, action);
    if (action.type.startsWith("DEVICE_CREATESTREAM_"))
        newState[action.name].create = streamCreateReducer(newState[action.name].create, action);

    return newState;
}

// get the device page from the state - the state might not have this
// particular page initialized, meaning that it wasn't acted upon
export function getDeviceState(device, state) {
    return (state.device[device] !== undefined
        ? state.device[device]
        : DeviceInitialState);
}
