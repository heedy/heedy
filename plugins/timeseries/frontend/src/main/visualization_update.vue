<template>
  <div>
    <h-not-found v-if="index === -1" />
    <visualization-editor
      v-else
      :visualization="currentVisualization"
      :index="index"
    />
  </div>
</template>
<script>
import VisualizationEditor from "./components/visualization_editor.vue";
export default {
  components: {
    VisualizationEditor,
  },
  props: {
    name: String,
  },
  head() {
    return {
      title: `Editing Visualization ${this.name}`,
    };
  },
  computed: {
    visualizations() {
      if (
        this.$store.state.app.info.settings.timeseries?.visualizations !==
        undefined
      ) {
        return this.$store.state.app.info.settings.timeseries.visualizations;
      }
      return [];
    },
    currentVisualization() {
      return this.visualizations[this.index];
    },
    index() {
      const name = this.name;
      console.log("EDITING",name)
      // Get the name of the visualization from the query
      for (let i = 0; i < this.visualizations.length; i++) {
        if (this.visualizations[i].name === name) {
          return i;
        }
      }

      return -1;
    },
  },
};
</script>
