// Actions are things that can happen... To make it happen, run store.dispatch(action())

// show the given page
export function showPage(page) {
    return {
        type: 'SHOW_PAGE',
        value: page
    };
}

// set the search bar text
export function setSearchText(text) {
    return {
        type: 'SET_SEARCH_TEXT',
        value: text
    };
}
