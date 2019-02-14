<template>
    <v-card>

        <v-card-title primary-title>
          <div>
            <div class="headline">ConnectorDB Setup</div>
            <span class="grey--text">Prepare the server's basic settings.</span>
          </div>
        </v-card-title>

        <v-card-actions>
            <v-btn icon @click="show = !show">
            <v-icon>{{ show ? 'keyboard_arrow_up' : 'keyboard_arrow_down' }}</v-icon>
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn flat>Go Back</v-btn>
          <v-btn flat color="purple">Continue</v-btn>
          
          
        </v-card-actions>

        <v-slide-y-transition>
          <v-card-text v-show="show">
            I'm a thing. But, like most politicians, he promised more than he could deliver. You won't have time for sleeping, soldier, not with all the bed making you'll be doing. Then we'll go with that data file! Hey, you add a one and two zeros to that or we walk! You're going to do his laundry? I've got to find a way to escape.
          </v-card-text>
        </v-slide-y-transition>
      </v-card>
</template>

<script>
export default {
  data: () => ({
      show: false
    }),
  props: {
  }
};
</script>
