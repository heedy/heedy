// The downlinks page is where you can control devices connected to ConnectorDB
import { DownlinkSearchInitialState, downlinkSearchReducer } from "./search";

export const DownlinkPageInitialState = {
  loaded: false,
  downlinks: [],

  search: DownlinkSearchInitialState
};

export default function downlinkPageReducer(state, action) {
  if (action.type.startsWith("DOWNLINK_SEARCH_"))
    return {
      ...state,
      search: downlinkSearchReducer(state.search, action)
    };

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
