<template>
    <div>
        <vue-headful :title="title"></vue-headful>
        <loading v-if="source==null"></loading>
        <router-view v-else :source="source"></router-view>
    </div>
</template>
<script>
import {Loading} from "../components.mjs";
export default {
    components: {
        Loading
    },
    data: () => ({}),
    props: {
        sourceid: String
    },
    watch: {
        sourceid(newValue) {
            this.$store.dispatch("readSource",{id: newValue});
        }
    },
    computed: {
        source() {
            return this.$store.state.heedy.sources[this.sourceid] || null;
        },
        title() {
            let s = this.source;
            if (s==null) {
                return "loading... | heedy";
            }
            return s.fullname + " | heedy";
        }
    },
    created() {
        this.$store.dispatch("readSource",{id: this.sourceid});
    }
}
</script>