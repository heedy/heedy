<template>
    <v-expansion-panels v-model="panel" accordion>
        <v-expansion-panel v-for="err in errors" :key="`${err.plugin}:${err.name}`">
            <v-expansion-panel-header>
                Error in the {{ err.name }} visualization
            </v-expansion-panel-header>
            <v-expansion-panel-content>
                <samp style="color:red">
                    {{ err.error }}
                </samp><br/>
                <samp style="color:gray;font-size:0.6em;line-height:1">
                    {{ err.stack }}
                </samp><br/>
                <div v-if="err.type=='user'" style="text-align: right">
                <v-btn v-if="err.name!==editingName" color="primary" text @click="editvis(err)">Edit Visualization</v-btn>
                </div>
                <p v-else><br/>This is a bug in the {{ err.name}} visualization! You should report it!</p>
            </v-expansion-panel-content>
        </v-expansion-panel>
    </v-expansion-panels>
</template>
<script>
export default {
    props: {
        query: Object,
        errors: Array,
        editingName: String,
    },
    data: () => ({
        panel: 0,
    }),
    methods: {
        editvis(err) {
            this.$router.push({
                path: `/timeseries/visualization/update/${encodeURIComponent(err.name)}`,
                query: {
                    test_query: btoa(JSON.stringify(this.query))
            }});
        }
    }
}
</script>
