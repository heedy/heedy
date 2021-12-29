<template>
  <div class="mt-2 mb-6 text-center" style="width: 100%">
    {{ label }}
    <v-rating
      :value="value"
      @input="(i) => $emit('input', i)"
      :length="10"
      color="yellow darken-3"
      background-color="grey darken-1"
      hover
      :size="size"
    ></v-rating>
  </div>
</template>
<script>
export default {
  props: {
    value: Number,
    label: String,
  },
  computed: {
    size() {
      // The size should really be computed based on the div width, but whatever, for now this hack will work
      if (this.$vuetify.breakpoint.name == "xs") {
        return 20;
      }
      if (this.$vuetify.breakpoint.name == "md") {
        return 24;
      }
      return 30;
    },
  },
};
</script>