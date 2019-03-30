<template>
  <v-card>
    <div style="position:absolute;bottom:2px;right:2px;">
      <v-btn icon>
        <v-icon style="color:lightgray">edit</v-icon>
      </v-btn>
    </div>
    <v-container grid-list-md>
      <v-layout row wrap>
        <v-flex xs12 sm4 md3 lg2 text-xs-center justify-center>
          <v-avatar :size="avatarSize" color="grey lighten-4">
            <img src="https://vuetifyjs.com/apple-touch-icon-180x180.png" alt="avatar">
          </v-avatar>
          <h5 style="color:gray;padding-top:10px">dkumor</h5>
        </v-flex>
        <v-flex xs12 sm8 md9 lg10>
          <h2>Daniel Kumor</h2>
          <p>I am a hooman</p>
        </v-flex>
      </v-layout>
    </v-container>
  </v-card>
</template>

<script>
export default {
  data: () => ({}),
  computed: {
    avatarSize() {
      return 120;
    }
  }
};
</script>