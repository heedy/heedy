<template>
    <v-expansion-panels v-model="panel" accordion>
        <v-expansion-panel v-for="err in data.errors" :key="`${err.plugin}:${err.name}`">
            <v-expansion-panel-header>
                Error in visualization {{ err.name }}
            </v-expansion-panel-header>
            <v-expansion-panel-content>
                <p>The visualization failed to process query data, giving the following error:</p>
                <samp style="color:red">
                    {{ err.error }}
                </samp>
                <p v-if="err.type=='plugin'">This is either a bug in Heedy, or in the plugin that defines the "{{ err.name }}" visualization.</p>
                <v-btn v-else color="primary" text @click="editvis(err)">Edit Visualization</v-btn>
            </v-expansion-panel-content>
        </v-expansion-panel>
    </v-expansion-panels>
</template>
<script>
export default {
    props: {
        query: Object,
        config: Object,
        data: Array,
        type: String,
    },
    data: () => ({
        panel: 0,
    }),
    methods: {
        editvis(err) {
            this.$router.push({
                path: "/timeseries/customize_visualization",
                query: {
                    name: err.name,
                    q: btoa(JSON.stringify(this.query))

            }});
        }
    }
}
</script>
