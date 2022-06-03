<template>
    <visualization-editor :visualization="defaultVisualization" />
</template>
<script>
import VisualizationEditor from "./components/visualization_editor.vue";
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
            code: "return vis;",
            name: "My Visualization",
            test_query: {"Series 1": {i1:-1000}}
        }
    }),
}
</script>
