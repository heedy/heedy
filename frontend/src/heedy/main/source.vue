<template>
    <component :is="is" :source="source" />
</template>
<script>

import {typeComponents} from "./sourceInjector.js";
import SourceDefault from "./source_default.vue";

export default {
    props: {
        source: Object
    },
    computed: {
        is() {
            return typeComponents[this.source.type] || SourceDefault;
        }
    }
}
</script>