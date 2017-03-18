// The analysis page is where you generate visualizations of datasets from your data
import moment from 'moment';

import { AnalysisSearchInitialState, analysisSearchReducer } from './search';


const DatasetStreamInitialState = {
    stream: "",
    transform: "",
    interpolator: "closest",
    allownil: false
};

export const AnalysisPageInitialState = {
    posttransform: "",
    transform: "",
    t1: moment().subtract(30, 'days'),
    t2: moment().endOf('day'),
    limit: 100000,
    stream: "",
    dataset: {
        y: DatasetStreamInitialState
    },
    error: null,
    views: {},
    loading: false,
    data: [],
    xdataset: true,
    dt: "60*60*24",

    search: AnalysisSearchInitialState
};

export default function AnalysisPageReducer(state, action) {
    if (action.type.startsWith("ANALYSIS_SEARCH_"))
        return {
            ...state,
            search: analysisSearchReducer(state.search, action)
        };

    let k = Object.keys(state.dataset).length;
    let d = {
        ...state.dataset
    }
    switch (action.type) {
        case "ANALYSIS_STATE":
            return {
                ...state,
                ...action.value,
                error: null
            };
        case "DATASET_STATE":
            d[action.key] = {
                ...state.dataset[action.key],
                ...action.value
            }
            return {
                ...state,
                dataset: d,
                error: null
            };
        case "ANALYSIS_CLEAR":
            return AnalysisPageInitialState;
        case "ANALYSIS_LOADING":
            return {
                ...state,
                loading: action.value,
            };
        case "ANALYSIS_ERROR":
            return {
                ...state,
                error: action.value
            };
        case "SHOW_DATASET":
            return {
                ...state,
                data: action.value,
                loading: false,
                error: null
            };
        case "ADD_DATASET_STREAM":
            // When adding a stream, we need to name it correctly. We go Z,A,B,C,...

            if (k == 1) {
                d["z"] = DatasetStreamInitialState;
            } else if (k < 10) {
                d[String.fromCharCode('a'.charCodeAt(0) + k - 2)] = DatasetStreamInitialState;
            }
            return {
                ...state,
                dataset: d
            };
        case "REMOVE_DATASET_STREAM":
            if (k < 2) {
                // do nothing
            } else if (k == 2) {
                delete d["z"];
            } else {
                delete d[String.fromCharCode('a'.charCodeAt(0) + k - 3)];
            }
            return {
                ...state,
                dataset: d
            };
    }
    return state;
}
