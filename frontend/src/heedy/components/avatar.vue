<template>
  <v-avatar :size="size" color="grey lighten-4">
    <v-icon
      v-if="image.startsWith('fa:') || image.startsWith('mi:')"
    >{{ image.substring(3,image.length) }}</v-icon>
    <img v-else-if="image.length > 0" :src="image">
    <v-icon v-else>person</v-icon>
  </v-avatar>
</template>

<script>
export default {
  props: {
    image: String,
    colorHash: String,
    size: {
      type: Number,
      default: 48
    }
  },
  computed: {
    /*
    imageStyle() {
      let s = { maxWidth: "100%", maxHeight: "100%" };
      if (this.width > 0) {
        s.width = this.width + "px";
        s.height = this.width + "px";
      }
      console.log(s);
      return s;
    }*/
  }
};
</script>