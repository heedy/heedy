// Actions are things that can happen... To make it happen, run store.dispatch(action())
import {push} from 'react-router-redux'

// set the search bar text
export function setSearchText(text) {
    return {type: 'SET_SEARCH_TEXT', value: text};
}
