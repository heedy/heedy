<template>
<v-app id="heedy">
    <Login v-if="user==null"></Login>
    <AllowAccess v-else ></AllowAccess>
</v-app>
</template>

<script>

// Login is imported from the path of the main auth component
import Login from "./heedy/login.mjs";

// AllowAccess is embedded in the bundle
import AllowAccess from "./oauth.vue";


export default {
    components:{
        Login,
        AllowAccess
    },
    computed: {
        user() {
            return this.$store.state.user;
        }
    }
};
</script>