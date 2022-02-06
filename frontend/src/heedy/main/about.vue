<template >
  <h-page-container>
    <v-card>
      <v-card-text class="text--primary">
        <v-container>
          <v-layout row justify-center>
            <v-flex text-center style="padding-top: 1cm">
              <h1 style="color: #1976d2">heedy</h1>
              <h4>Live an examined life.</h4>
            </v-flex>
          </v-layout>
          <v-layout row justify-center>
            <v-flex
              text-center
              xs12
              md8
              lg6
              style="padding-top: 1cm; backgroud-color: blue"
            >
              <p>
                Heedy is an open-source, extensible home for your personal data.
                It enables aggregating and analyzing the data your devices and
                apps gather about you.
              </p>
            </v-flex>
          </v-layout>
          <v-layout row justify-center>
            <v-flex text-center style="padding-top: 1cm; backgroud-color: blue">
              <h5 style="padding-bottom: 1cm">
                For more information visit
                <a href="https://heedy.org">heedy.org</a>
              </h5>
            </v-flex>
          </v-layout>
        </v-container>
      </v-card-text>
    </v-card>
  </h-page-container>
</template>

<script>
export default {
  head: {
    title: "About",
  },
};
</script>