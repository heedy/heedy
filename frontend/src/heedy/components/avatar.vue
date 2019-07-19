<template>
  <v-avatar :size="size" :color="color">
    <v-icon
      v-if="image.startsWith('fa:') || image.startsWith('mi:')"
    >{{ image.substring(3,image.length) }}</v-icon>
    <img v-else-if="image.length > 0" :src="image">
    <v-icon v-else :size="iconSize">{{ defaultIcon }}</v-icon>
  </v-avatar>
</template>

<script>
import ColorHash from "color-hash";

var colorHash = new ColorHash();

export default {
  props: {
    image: String,
    colorHash: String,
    size: {
      type: Number,
      default: 48
    },
    defaultIcon: {
      type: String,
      default: "person"
    }
  },
  computed: {
    color() {
      return colorHash.hex(this.colorHash);
    },
    iconSize() {
      return Math.round(0.7 * this.size);
    }
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