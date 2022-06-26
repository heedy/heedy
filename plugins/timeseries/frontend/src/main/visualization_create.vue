<template>
    <visualization-editor :visualization="defaultVisualization" />
</template>
<script>
import VisualizationEditor from "./components/visualization_editor.vue";

const defaultCode = `// Check if the data can be visualized
if (c.data.length!=1) return; // Multiple series
if (c.data[0].length < 5) return; // Less than 5 datapoints

const d = c.data[0].d;
if (d.type()!=="number") return; // Not a numeric timeseries

// If it can, create a visualization
vis.set("my_visualization",{
    title: "My Visualization",
    weight: 15,
    type: "??", // Set a visualization type
    config: {
        // Add configuration for the visualization type
    }
});`
export default {
    components: {
        VisualizationEditor,
    },
    head: {
        title: "Create New Visualization",
    
    },
    data: ()=> ({
        defaultVisualization: {
            enabled: true,
            code: defaultCode,
            name: "My Visualization",
            test_query: {"Series 1": {i1:-1000}}
        }
    }),
}
</script>
