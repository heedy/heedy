// The downlinks page is where you can control devices connected to ConnectorDB

export const DownlinkPageInitialState = {
    loaded: false,
    downlinks: []
};

export default function downlinkPageReducer(state, action) {
    switch (action.type) {
        case "UPDATE_DOWNLINKS":
            return {
                ...state,
                loaded: true,
                downlinks: action.value
            };
    }
    return state;
}
