<template>
  <v-app id="inspire">
  <v-content>
  
  <v-container fluid fill-height>
    <v-layout justify-center align-center>
      <v-flex xs12 lg6 md8>
        <router-view/>
      </v-flex>
    </v-layout>
  </v-container>
</v-content>

  </v-app>
</template>

<script>
export default {
  data: () => ({
    drawer: true,
    items: [
      { title: "Home", icon: "dashboard" },
      { title: "About", icon: "question_answer" }
    ],
    mini: null
  }),
  props: {
    object: String
  }
};
</script>

<style>
#inspire {
  background: linear-gradient(to bottom, #182447, #215D85);
}
</style>