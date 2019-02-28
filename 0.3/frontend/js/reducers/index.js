import userReducer from "./user";
import deviceReducer from "./device";
import streamReducer from "./stream";
import siteReducer from "./site";
import pageReducer from "./pages";

export const reducers = {
  user: userReducer,
  device: deviceReducer,
  stream: streamReducer,
  site: siteReducer,
  pages: pageReducer
};
