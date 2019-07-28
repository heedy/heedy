<template>
    <div>
        <loading v-if="source==null"></loading>
        <h1 v-else>{{ source.name }}</h1>
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
        }
    },
    created() {
        console.log(this.sourceid);
        this.$store.dispatch("readSource",{id: this.sourceid});
    }
}
</script>