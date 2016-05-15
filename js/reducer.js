import InitialState, {UserPageInitialState} from './state';

function userPageReducer(state = UserPageInitialState, action) {
    switch (action.type) {
        case 'USERPAGE_EDIT':
            return {
                ...state,
                editing: action.value
            }
        case 'USERPAGE_EXPAND':
            return {
                ...state,
                expanded: action.value
            }
    }
    return state;
}

export default function reducer(state = InitialState, action) {
    switch (action.type) {
        case 'LOAD_CONTEXT':
            var out = Object.assign({}, state, {
                siteURL: action.value.SiteURL,
                thisUserName: action.value.ThisUser.name,
                thisDeviceName: action.value.ThisDevice.name
            });

            // now set up the navigation correctly (replace {self} with user name)
            for (var i = 0; i < out.navigation.length; i++) {
                out.navigation[i].title = out.navigation[i].title.replace("{self}", out.thisUserName);
                out.navigation[i].subtitle = out.navigation[i].subtitle.replace("{self}", out.thisUserName);
                out.navigation[i].page = out.navigation[i].page.replace("{self}", out.thisUserName);
            }
            return out;

        case 'SET_SEARCH_TEXT':
            return {
                ...state,
                searchText: action.value
            };
    }
    if (action.type.startsWith("USERPAGE")) {
        let newpage = {
            ...state,
            userpage: {
                ...state.userpage
            }
        };
        newpage.userpage[action.name] = userPageReducer(newpage.userpage[action.name], action);
        return newpage;
    }

    return state;
}
