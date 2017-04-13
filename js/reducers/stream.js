// The streamReducer maintains state for all streams - each stream page that was visited in this session has
// its state maintained here. Each action that operates on a stream contains a "name" field, which states the stream
// in whose context to operate

// The keys in this object are going to be the stream names
const InitialState = {};

import streamViewReducer, { StreamViewInitialState } from "./streamView";
import streamEditReducer, { StreamEditInitialState } from "./streamEdit";

export const StreamInputInitialState = {
  expanded: false
};

// The initial state of a specific stream
const StreamInitialState = {
  edit: StreamEditInitialState,
  view: StreamViewInitialState,
  input: StreamInputInitialState
};

export default function streamReducer(state = InitialState, action) {
  if (!action.type.startsWith("STREAM_")) return state;

  // Set up the new state
  let newState = {
    ...state
  };

  // If the stream already has a state, copy the next level, otherwise, initialize the next level
  if (state[action.name] !== undefined) {
    newState[action.name] = {
      ...state[action.name]
    };
  } else {
    newState[action.name] = Object.assign({}, StreamInitialState);
  }

  // Now route to the appropriate reducer
  if (action.type.startsWith("STREAM_EDIT")) {
    newState[action.name].edit = streamEditReducer(
      newState[action.name].edit,
      action
    );
  } else if (action.type.startsWith("STREAM_VIEW_")) {
    newState[action.name].view = streamViewReducer(
      newState[action.name].view,
      action
    );
  } else if (action.type == "STREAM_INPUT") {
    newState[action.name].input = action.value;
  } else if (action.type == "STREAM_CLEAR_STATE") {
    newState[action.name] = StreamInitialState;
  }
  return newState;
}

// get the stream page from the state - the state might not have this
// particular page initialized, meaning that it wasn't acted upon
export function getStreamState(stream, state) {
  return state.stream[stream] !== undefined
    ? state.stream[stream]
    : StreamInitialState;
}
