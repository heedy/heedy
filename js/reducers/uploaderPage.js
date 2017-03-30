// The downlinks page is where you can control devices connected to ConnectorDB
import { UploaderSearchInitialState, uploaderSearchReducer } from './search';

export const UploaderPageInitialState = {
    search: UploaderSearchInitialState,
    part1: {
        width: "half",
        rawdata: "Paste your data here..."
    },
    part2: {
        width: "half",
        transform: "",
        error: ""
    },
    part3: {
        stream: "",
        create: true,
        overwrite: false,
        removeolder: false,
        loading: false,
        percentdone: 0,
        error: ""
    },

    data: []
};

export default function downlinkPageReducer(state, action) {
    if (action.type.startsWith("UPLOADER_SEARCH_"))
        return {
            ...state,
            search: uploaderSearchReducer(state.search, action)
        };

    switch (action.type) {
        case "UPLOADER_PART1":
            return {
                ...state,
                part1: {
                    ...state.part1,
                    ...action.value
                }
            };
        case "UPLOADER_PART2":
            return {
                ...state,
                part2: {
                    ...state.part2,
                    ...action.value
                }
            };
        case "UPLOADER_PART3":
            return {
                ...state,
                part3: {
                    ...state.part3,
                    ...action.value
                }
            };
        case "UPLOADER_SET":
            return {
                ...state,
                ...action.value
            };
    }
    return state;
}
