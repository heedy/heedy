<template>
  <v-main v-resize="onResize">
    <v-container grid-list-xl :fluid="fluid">
      <v-layout fill-height column>
        <v-flex v-if="!noflex">
          <slot></slot>
        </v-flex>
        <slot v-else></slot>
      </v-layout>
    </v-container>
  </v-main>
</template>
<script>
export default {
  data: () => ({
    fluid: true,
  }),
  props: {
    noflex: {
      type: Boolean,
      default: false,
    },
  },
  mounted() {
    this.onResize();
  },
  methods: {
    onResize() {
      this.fluid = window.innerWidth < 1264;
    },
  },
};
</script>
