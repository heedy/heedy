<template>
    <div>
        <template v-if="iconMode">
             <avatar :size="size-30" defaultIcon="person" :image="iconText" >
            </avatar><br/>
            <v-text-field class="centered-input" label="Icon Name" placeholder="person" v-model="iconText"></v-text-field>
            <v-btn small text @click="iconMode = false" >Custom Image</v-btn>
        </template>
        <template v-else>
            <croppa :width="size" :height="size" ref="imageCropper"></croppa><br/>
            <v-btn small text @click="iconMode = true" >Font Icons</v-btn>
        </template>
    </div>
</template>
<script>
import Croppa from "vue-croppa";
import "vue-croppa/dist/vue-croppa.css";

import Avatar from "./avatar.vue";

export default {
    components: {
        Croppa: Croppa.component,
        Avatar
    },
    data: () => ({
        iconMode: false,
        iconText: "",
    }),
    props: {
        image: String,
        size: {
            default: 160,
            type: Number
        }
    },
    watch: {
        image: {
            immediate: true,
            handler(newImage,oldImage) {
                let iconMode = !newImage.startsWith("data:image/") ;
                let iconText = "";
                if (iconMode) {
                    iconText = this.image;
                }

                this.iconMode = iconMode;
                this.iconText = iconText;

            }
        }
    },
    methods: {
        getImage() {
            if (this.iconMode) {
                return this.iconText;
            }
            if (!this.$refs.imageCropper.hasImage()) {
                return this.image;
            }
            return this.$refs.imageCropper.generateDataUrl();
        },
        hasImage() {
            if (this.iconMode) {
                if (this.iconText=="") {
                    return false;
                }
                return (this.iconText != this.image);
            }
            return this.$refs.imageCropper.hasImage()
        }
    }
}
</script>
<style>
.centered-input input {
  text-align: center
}
</style>