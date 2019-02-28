import { UserSearchInitialState, userSearchReducer } from "./search";

export const UserViewInitialState = {
  expanded: true,
  hidden: true,
  search: UserSearchInitialState
};

export default function userViewReducer(state, action) {
  if (action.type.startsWith("USER_VIEW_SEARCH_"))
    return {
      ...state,
      search: userSearchReducer(state.search, action)
    };

  switch (action.type) {
    case "USER_VIEW_EXPANDED":
      return {
        ...state,
        expanded: action.value
      };
    case "USER_VIEW_HIDDEN":
      return {
        ...state,
        hidden: action.value
      };
  }
  return state;
}
