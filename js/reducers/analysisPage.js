// The analysis page is where you generate visualizations of datasets from your data
import moment from 'moment';

export const AnalysisPageInitialState = {
    posttransform: "",
    transform: "",
    t1: moment().subtract(7, 'days'),
    t2: moment().endOf('day'),
    limit: 0,
    stream: "",
    dataset: {
        y: {
            stream: "",
            transform: "",
            interpolator: "closest",
            allownil: false
        }
    },
    error: null,
    views: {},
    loading: false,
    data: []
};

export default function AnalysisPageReducer(state, action) {
    switch (action.type) {
        case "ANALYSIS_STATE":
            return {
                ...state,
                ...action.value,
                error: null
            };
        case "DATASET_STATE":
            let datasetobj = {
                ...state.dataset,
            };
            datasetobj[action.key] = {
                ...state.dataset[action.key],
                ...action.value
            }
            return {
                ...state,
                dataset: datasetobj,
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
                loading: false
            };
    }
    return state;
}
