<template >
  <v-content>
    <v-container grid-list-xl>
      <v-layout fill-height column>
        <v-flex v-if="!noflex">
          <slot></slot>
        </v-flex>
        <slot v-else></slot>
      </v-layout>
    </v-container>
  </v-content>
</template>
<script>
export default {
  props: {
    noflex: {
      type: Boolean,
      default: false
    }
  }
};
</script>