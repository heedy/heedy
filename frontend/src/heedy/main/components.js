/**
 * This file contains the main components that are used in Heedy's core.
 * The components here will be stable over minor versions, so plugins
 * are free to use them for their own UI
 */

import Loading from "./components/loading.vue";
import Avatar from "./components/avatar.vue";
import AvatarEditor from "./components/avatar_editor.vue";
import ScopeEditor from "./components/scope_editor.vue";
import PageContainer from "./components/page_container.vue";
import CardPage from "./components/card_page.vue";
import NotFound from "./components/404.vue";
import Header from "./components/header.vue";

function register(Vue) {
    Vue.component("h-loading", Loading);
    Vue.component("h-avatar", Avatar);
    Vue.component("h-avatar-editor", AvatarEditor);
    Vue.component("h-scope-editor", ScopeEditor);
    Vue.component("h-page-container", PageContainer);
    Vue.component("h-card-page", CardPage);
    Vue.component("h-not-found", NotFound);
    Vue.component("h-header", Header)
}

export default register;