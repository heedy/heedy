<template>
  <v-avatar :size="size" :color="color">
    <img v-if="image.startsWith('data:image/')" :src="image" />
    <v-icon
      v-else-if="image.length > 0"
      :size="iconSize"
      :style="{'fontSize': fontSize}"
    >{{ image }}</v-icon>
    <v-icon v-else :size="iconSize" :style="{'fontSize': fontSize}">{{ defaultIcon }}</v-icon>
  </v-avatar>
</template>

<script>
import ColorHash from "color-hash";

var colorHash = new ColorHash();

export default {
  props: {
    image: {
      type: String,
      default: ""
    },
    colorHash: String,
    size: {
      type: Number,
      default: 48
    },
    defaultIcon: {
      type: String,
      default: "brightness_1"
    }
  },
  computed: {
    color() {
      return colorHash.hex(this.colorHash);
    },
    iconSize() {
      return Math.round(0.7 * this.size);
    },
    fontSize() {
      if (this.image.includes("fa-")) {
        return Math.round(0.65 * this.iconSize) + "px";
      }
      return Math.round(0.9 * this.iconSize) + "px";
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