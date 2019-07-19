/**
 * This file contains the main components that are used in Heedy's core.
 * The components here will be stable over minor versions, so plugins
 * are free to use them for their own UI
 */

import Loading from "./components/loading.vue";
import Avatar from  "./components/avatar.vue";
import PageContainer from "./components/page_container.vue";
import NotFound from "./components/404.vue";

export {Loading, NotFound, Avatar, PageContainer};