import userReducer from './user';
import deviceReducer from './device';
import streamReducer from './stream';
import queryReducer from './query';
import siteReducer from './site';

export const reducers = {
    user: userReducer,
    device: deviceReducer,
    stream: streamReducer,
    site: siteReducer,
    query: queryReducer
};
