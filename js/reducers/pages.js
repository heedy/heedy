// The pages reducer holds the state for all of the non- user/device/stream pages,
// such as the insert page, the explore page, etc.

import indexPageReducer, { IndexPageInitialState } from './indexPage';
import downlinkPageReducer, { DownlinkPageInitialState } from './downlinksPage';
import analysisPageReducer, { AnalysisPageInitialState } from './analysisPage';

const InitialState = {
    index: IndexPageInitialState,
    downlinks: DownlinkPageInitialState,
    analysis: AnalysisPageInitialState
}

export default function pageReducer(state = InitialState, action) {
    if (!action.type.startsWith("PAGE_"))
        return state;

    // Set up the new state
    return {
        index: indexPageReducer(state.index, action),
        downlinks: downlinkPageReducer(state.downlinks, action),
        analysis: analysisPageReducer(state.analysis, action)
    };
}
