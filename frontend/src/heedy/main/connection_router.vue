<template>
    <div>
        <loading v-if="loading"></loading>
        <not-found v-else-if="connection==null" />
        <router-view v-else :connection="connection"></router-view>
    </div>
</template>
<script>
import {Loading, NotFound} from "../components.mjs";
export default {
    components: {
        Loading, NotFound
    },
    data: () => ({}),
    props: {
        connectionid: String
    },
    watch: { 
        connectionid(newValue) {
            this.$store.dispatch("readConnection", { id: newValue });
        }
    },
    computed: {
        connection() {
            let c = this.$store.state.heedy.connections[this.connectionid] || null;
            return c;
        },
        loading() {
            return this.$store.state.heedy.connections == null;
        }
    },
    created() {
        this.$store.dispatch("readConnection", { id: this.connectionid });
    }
}
</script>