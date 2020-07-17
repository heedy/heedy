<template>
  <h-object-header :object="object">
    <range-picker
      style="padding-top: 17px;padding-right: 10px; max-width: 280px;"
      v-model="query"
    ></range-picker>
  </h-object-header>
</template>
<script>
import RangePicker from "./dataset/range_picker.vue";
export default {
  components: {
    RangePicker,
  },
  props: {
    object: Object,
  },
  data: () => ({
    dialog: false,
    live: true,
  }),
  computed: {
    query: {
      get() {
        return this.$route.query;
      },
      set(v) {
        this.$router.replace({ query: v });
      },
    },
  },
};
</script>
