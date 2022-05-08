<template>
  <div style="padding: 20px">
    <h5>Visualizations</h5>
    <p>
      You can customize the plots and information that is shown automatically
      for different datasets here by setting up custom JavaScript code that
      decides when and how to show visualizations. You can alter existing
      visualizations or create new ones. The custom code is run in order, so a
      visualization has access to and can modify the output of the previous one.
    </p>
    <v-simple-table
      v-if="value.visualizations.length > 0"
      fixed-header
      
    >
        <thead>
          <tr>
            <th class="text-left">Name</th>
            <th class="text-right">Enabled</th>
          </tr>
        </thead>
      
      <tbody>
        <tr v-for="(vis,idx) in value.visualizations" :key="vis.name" >
          <td @click="editVis(vis.name)" style="cursor: pointer">{{ vis.name }}</td>
          <td class="text-right" style="width:100px">
            <v-switch :input-value="vis.enabled" :true-value="true" @change="(v)=> setVis(idx,v)" style="float:right"/>
          </td>
        </tr>
      </tbody>
    </v-simple-table>
    <div
      :style="{
        width: '100%',
        'padding-top': '10px',
        'text-align': value.visualizations.length > 0 ? 'right' : 'center',
      }"
    >
      <v-btn
        class="mx-2"
        fab
        dark
        color="primary"
        to="/timeseries/customize_visualization"
      >
        <v-icon dark>add</v-icon>
      </v-btn>
    </div>
  </div>
</template>
<script>
export default {
  props: {
    schema: Object,
    value: Object,
    plugin: String,
  },
  methods: {
    editVis(name) {
      this.$router.push({path:"/timeseries/customize_visualization", query: {name: name}});
    },
    setVis(idx,v) {
      const visualizations = [...this.value.visualizations];
      visualizations[idx] = {...visualizations[idx], enabled: v};
      this.$emit("update",{
        visualizations
      });
    }
  }
};
</script>