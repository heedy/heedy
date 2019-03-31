<template >
  <page-container>
    <v-layout row justify-center>
      <v-flex justify-center text-xs-center xs12 style="padding-top: 1cm;">
        <h1 style="color:#1976d2">heedy</h1>
        <h4>Live an examined life.</h4>
      </v-flex>
    </v-layout>
    <v-layout row justify-center>
      <v-flex text-xs-center xs12 md8 lg6 style="padding-top: 1cm;backgroud-color:blue;">
        <p>
          Heedy is an open-source, extensible home for your quantified-self data,
          and other information your devices gather about you.
          It enables you easily analyze and share
          things you find in your data, and integrates with the most popular data gathering
          services.
        </p>
        <p>
          Heedy also contains a control system, allowing you to send commands to smart-home appliances
          based on your data.
        </p>
      </v-flex>
    </v-layout>
    <v-layout row justify-center>
      <v-flex text-xs-center xs12 md8 lg6 style="padding-top: 1cm;backgroud-color:blue;">
        <h5 style="padding-bottom: 1cm">
          For more information visit
          <a hfre="https://heedy.org">heedy.org</a>
        </h5>
      </v-flex>
    </v-layout>
  </page-container>
</template>

<script>
import PageContainer from "../embedded/page_container.vue";

export default {
  components: {
    PageContainer
  }
};
</script>