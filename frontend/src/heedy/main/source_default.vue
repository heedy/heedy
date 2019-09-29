<template>
  <h-page-container>
    <v-card>
      <v-container grid-list-md>
        <v-layout row wrap>
          <v-flex xs12 sm4 md3 lg2 text-center justify-center>
            <h-avatar
              :size="120"
              :image="source.avatar"
              defaultIcon="assignment"
              :colorHash="source.name"
            ></h-avatar>
            <h5 style="color:gray;padding-top:10px">{{source.name}}</h5>
          </v-flex>
          <v-flex xs12 sm8 md9 lg10>
            <h2>{{ source.fullname==""?source.name:source.fullname}}</h2>
            <p v-if="source.description!=''">{{ source.description }}</p>
            <p v-else style="color:lightgray;">No description given.</p>
          </v-flex>
          <v-flex md12>
            <codemirror :options="cmOptions" :value="JSON.stringify(source,null,'  ')" />
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
  </h-page-container>
</template>
<script>
export default {
  props: {
    source: Object
  },
  data: () => ({
    cmOptions: {
      readOnly: true
    }
  })
};
</script>
<style>
.CodeMirror {
  height: auto;
}
</style>
