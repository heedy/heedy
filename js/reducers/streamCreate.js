export const StreamCreateInitialState = {};

export default function deviceCreateReducer(state, action) {
  switch (action.type) {
    case "DEVICE_CREATESTREAM_CLEAR":
      return StreamCreateInitialState;
    case "DEVICE_CREATESTREAM_SET":
      return Object.assign({}, state, action.value);
    case "DEVICE_CREATESTREAM_NAME":
      return {
        ...state,
        name: action.value
      };
    case "DEVICE_CREATESTREAM_SCHEMA":
      return {
        ...state,
        schema: action.value
      };
    case "DEVICE_CREATESTREAM_NICKNAME":
      return {
        ...state,
        nickname: action.value
      };
    case "DEVICE_CREATESTREAM_DESCRIPTION":
      return {
        ...state,
        description: action.value
      };
    case "DEVICE_CREATESTREAM_DOWNLINK":
      return {
        ...state,
        downlink: action.value
      };
    case "DEVICE_CREATESTREAM_EPHEMERAL":
      return {
        ...state,
        ephemeral: action.value
      };
    case "DEVICE_CREATESTREAM_DATATYPE":
      return {
        ...state,
        datatype: action.value
      };
  }
  return state;
}
