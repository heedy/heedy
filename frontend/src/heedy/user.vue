<template >
  <v-content>
    <v-container fill-height>
      <v-layout>
        <v-flex>
          <user></user>
        </v-flex>
      </v-layout>
    </v-container>
  </v-content>
</template>

<script>
import User from "./components/user.mjs";
export default {
  components: {
    User
  },
  data: () => ({}),
  props: {
    source: String
  }
};
</script>

<style>
</style>