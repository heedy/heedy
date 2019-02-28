import { DeviceSearchInitialState, deviceSearchReducer } from "./search";

export const DeviceViewInitialState = {
  expanded: true,
  search: DeviceSearchInitialState
};

export default function deviceViewReducer(state, action) {
  if (action.type.startsWith("DEVICE_VIEW_SEARCH_"))
    return {
      ...state,
      search: deviceSearchReducer(state.search, action)
    };

  switch (action.type) {
    case "DEVICE_VIEW_EXPANDED":
      return {
        ...state,
        expanded: action.value
      };
  }
  return state;
}
