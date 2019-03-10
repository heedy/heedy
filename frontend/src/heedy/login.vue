<template>
  <v-content class="login-background">
    <v-container fluid>
      <v-layout justify-center align-center>
        <v-flex text-xs-center>
          <v-card class="mx-auto" max-width="400">
            <v-card-title>
              <span class="title font-weight-light">Log In</span>
            </v-card-title>
            <v-card-text class="headline font-weight-bold">
              <v-text-field prepend-icon="person" name="Username" label="Username"></v-text-field>
              <v-text-field prepend-icon="lock" name="Password" label="Password" type="password"></v-text-field>
            </v-card-text>

            <v-card-actions>
              <v-btn primary large block>Login</v-btn>
            </v-card-actions>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </v-content>
</template>

<script>
export default {
  data: () => ({}),
  props: {
    source: String
  }
};
</script>

<style>
.login-background {
  /*background: linear-gradient(to bottom, #182447, #215D85);*/
  margin-top: 10%;
}
</style>