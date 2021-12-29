<template>
  <h-object-header :object="object">
    <h-timeseries-range-picker
      :style="rangeStyle"
      v-model="query"
      :icon="!$vuetify.breakpoint.xs"
    ></h-timeseries-range-picker>
  </h-object-header>
</template>
<script>
export default {
  props: {
    object: Object,
  },
  data: () => ({
    dialog: false,
    live: true,
    windowWidth: window.innerWidth,
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
    rangeStyle() {
      let o = {
        paddingTop: this.$vuetify.breakpoint.xs ? "17px" : "15px",
        paddingRight: "10px",
        fontSize: "70%",
      };
      if (this.$vuetify.breakpoint.xs) {
        // Get the screen size, we have to do this manually, idk why it isn't working normally...

        o.width = this.windowWidth - 200 + "px";
      } else {
        o.maxWidth = "350px";
      }
      return o;
    },
  },
  methods: {
    onResize(event) {
      this.windowWidth = window.innerWidth;
    },
  },
  mounted() {
    window.addEventListener("resize", this.onResize);
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.onResize);
  },
};
</script>
