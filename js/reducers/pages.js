// The pages reducer holds the state for all of the non- user/device/stream pages,
// such as the insert page, the explore page, etc.

import indexPageReducer, {IndexPageInitialState} from './indexPage';

const InitialState = {
    index: IndexPageInitialState
}

export default function pageReducer(state = InitialState, action) {
    if (!action.type.startsWith("PAGE_"))
        return state;

    // Set up the new state
    let newState = {
        ...state
    };

    // Now route to the appropriate reducers
    if (action.type.startsWith("PAGE_INDEX_"))
        newState.index = indexPageReducer(newState.index, action);

    return newState;
}
