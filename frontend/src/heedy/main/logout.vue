<template>
  <v-main class="loading">
    <v-container fluid>
      <v-layout justify-center align-center>
        <v-flex text-center>
          <h1>Logging Out...</h1>
        </v-flex>
      </v-layout>
    </v-container>
  </v-main>
</template>

<script>
export default {
  mounted: function() {
    console.log("Logging out");
    location.href = location.href.split("#")[0] + "auth/logout";
  },
};
</script>

<style>
.loading {
  color: #c9c9c9;
  margin-top: 10%;
}
</style>
