<template>
  <Login v-if="user == null" />
  <v-main v-else class="notfound">
    <v-container fluid>
      <v-layout justify-center align-center>
        <v-flex text-center>
          <h1>404</h1>
          <p>The requested resource is not available</p>
        </v-flex>
      </v-layout>
    </v-container>
  </v-main>
</template>
<script>
import Login from "../login.vue";

export default {
  components: {
    Login,
  },
  computed: {
    user() {
      return this.$store.state.app.info.user;
    },
  },
};
</script>

<style>
.notfound {
  color: #c9c9c9;
  margin-top: 10%;
}
</style>
