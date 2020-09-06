<template>
  <v-app id="inspire">
    <v-main>
      <v-container fluid fill-height>
        <v-layout justify-center align-center>
          <v-flex xs12 lg6 md8>
            <router-view />
          </v-flex>
        </v-layout>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
export default {
  data: () => ({}),
};
</script>

<style>
#inspire {
  background: linear-gradient(to bottom, #182447, #215d85);
}
</style>
