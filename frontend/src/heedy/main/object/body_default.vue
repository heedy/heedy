<template>
  <v-flex>
    <v-card>
      <v-container grid-list-md>
        <v-layout row wrap>
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
  })
};
</script>
<style>
.CodeMirror {
  height: auto;
}
</style>
