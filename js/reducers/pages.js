// The pages reducer holds the state for all of the non- user/device/stream pages,
// such as the insert page, the explore page, etc.

import indexPageReducer, { IndexPageInitialState } from "./indexPage";
import downlinkPageReducer, { DownlinkPageInitialState } from "./downlinksPage";
import analysisPageReducer, { AnalysisPageInitialState } from "./analysisPage";
import uploaderPageReducer, { UploaderPageInitialState } from "./uploaderPage";

const InitialState = {
  index: IndexPageInitialState,
  downlinks: DownlinkPageInitialState,
  analysis: AnalysisPageInitialState,
  uploader: UploaderPageInitialState
};

export default function pageReducer(state = InitialState, action) {
  // Set up the new state
  return {
    index: indexPageReducer(state.index, action),
    downlinks: downlinkPageReducer(state.downlinks, action),
    analysis: analysisPageReducer(state.analysis, action),
    uploader: uploaderPageReducer(state.uploader, action)
  };
}
