import InitialState from './state';

export default function reducer(state = InitialState, action) {
    console.log(action);
    switch (action.type) {
        case 'LOAD_CONTEXT':
            var out = Object.assign({}, state, {context: action.value});
            // now set up the navigation correctly (replace {self} with user name)
            for (var i = 0; i < out.navigation.length; i++) {
                out.navigation[i].title = out.navigation[i].title.replace("{self}", out.context.ThisUser.name);
                out.navigation[i].subtitle = out.navigation[i].subtitle.replace("{self}", out.context.ThisUser.name);
                out.navigation[i].page = out.navigation[i].page.replace("{self}", out.context.ThisUser.name);
            }
            return out;

        case 'SHOW_PAGE':
            return {
                ...state,
                linkSelected: action.value
            };

        case 'SET_SEARCH_TEXT':
            return {
                ...state,
                searchText: action.value
            };
    }
    return state;
}
