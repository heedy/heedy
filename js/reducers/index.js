import userReducer from './user';
import queryReducer from './query';
import siteReducer from './site';

export const reducers = {
    user: userReducer,
    site: siteReducer,
    query: queryReducer
};
