<template>
  <v-flex>
    <v-card>
      <v-container grid-list-md>
        <v-layout row wrap>
          <v-flex xs12 sm4 md3 lg2 text-center justify-center>
            <h-icon
              :size="120"
              :image="object.icon"
              :defaultIcon="defaultIcon"
              :colorHash="object.id"
            ></h-icon>
            <h5 style="color:gray;padding-top:10px">{{object.type}}</h5>
          </v-flex>
          <v-flex xs12 sm8 md9 lg10>
            <h2>{{ object.name }}</h2>
            <p v-if="object.description!=''">{{ object.description }}</p>
            <p v-else style="color:lightgray;">No description given.</p>
          </v-flex>
          <v-flex md12>
            <codemirror :options="cmOptions" :value="JSON.stringify(object,null,'  ')" />
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
  </v-flex>
</template>
<script>
export default {
  props: {
    object: Object
  },
  data: () => ({
    cmOptions: {
      readOnly: true
    }
  }),
  computed: {
    defaultIcon() {
      let otype = this.$store.state.heedy.object_types[this.object.type] || {
        icon: "assignment"
      };
      return otype.icon;
    }
  }
};
</script>
<style>
.CodeMirror {
  height: auto;
}
</style>
