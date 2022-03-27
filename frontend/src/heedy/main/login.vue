<template>
  <v-main class="login-background">
    <v-container fluid>
      <v-layout justify-center align-center>
        <v-flex text-center>
          <v-card class="mx-auto" max-width="400">
            <form @submit.prevent="login">
              <v-card-title>
                <span class="title font-weight-light">Log In</span>
              </v-card-title>
              <v-card-text class="headline font-weight-bold">
                <v-text-field
                  prepend-icon="person"
                  name="Username"
                  label="Username"
                  v-model="username"
                  autofocus
                  autocomplete="username"
                ></v-text-field>
                <v-text-field
                  prepend-icon="lock"
                  name="Password"
                  label="Password"
                  v-model="password"
                  type="password"
                  autocomplete="current-password"
                ></v-text-field>
              </v-card-text>

              <v-card-actions>
                <v-btn primary large block :loading="loading" type="submit"
                  >Login</v-btn
                >
              </v-card-actions>
            </form>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </v-main>
</template>

<script>
import api from "../../util.mjs";
export default {
  data: () => ({
    loading: false,
    username: "",
    password: "",
  }),
  head: {
    title: "Log In",
  },
  methods: {
    login: async function (e) {
      console.vlog("run login");
      this.loading = true;
      let result = await api(
        "POST",
        "auth/token",
        {
          grant_type: "password",
          username: this.username,
          password: this.password,
        },
        null,
        false
      );
      this.loading = false;
      if (!result.response.ok) {
        this.$store.dispatch("errnotify", result.data);
        this.password = "";
      } else {
        let locsplit = window.location.href.split("#");

        // Success, so perform a refresh of the page
        if (locsplit.length == 2 && locsplit[1] == "/login") {
          // If at login page, go to root
          window.location.href = locsplit[0];
        } else {
          // If elsewhere, move back there
          window.location.reload(true);
        }
      }
    },
  },
};
</script>

<style>
.login-background {
  /*background: linear-gradient(to bottom, #182447, #215D85);*/
  margin-top: 10%;
}
</style>
