// The IndexPage is the "insert" page where users are asked to manually insert Data
// such as their ratings/diaries/etc

import {IndexSearchInitialState, indexSearchReducer} from './search';

export const IndexPageInitialState = {
    search: IndexSearchInitialState
};

export default function indexPageReducer(state, action) {
    if (action.type.startsWith("PAGE_INDEX_SEARCH_")) {
        return {
            ...state,
            search: indexSearchReducer(state.search, action)
        };
    }
    /*
  switch (action.type) {
    case "PAGE_INDEX_SEARCH_"
  }
  */
}
