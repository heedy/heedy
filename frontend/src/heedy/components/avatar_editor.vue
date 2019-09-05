<template>
    <v-container >
        <v-layout column align-center style="border: 1px solid; border-radius: 4px; padding: 6px; border-color: #F0F0F0;">
        <template v-if="iconMode">
             <avatar :size="size-30" defaultIcon="person" :colorHash="colorHash" :image="iconText" ></avatar>
            <v-text-field class="centered-input" label="Icon Name" placeholder="person" v-model="iconText"></v-text-field>
            <a href="https://material.io/resources/icons/?style=baseline" target="_blank" style="font-size: 70%; margin-top: -15px; margin-bottom: 10px; color: gray; z-index: 1;">See available icons</a>
            <v-btn small text @click="iconMode = false" >Custom Image</v-btn>
        </template>
        <template v-else>
            <v-flex style="margin-bottom: 5px">
            <croppa :width="size" :height="size" ref="imageCropper"></croppa>
            </v-flex>
            <v-btn small text @click="iconMode = true">Font Icons</v-btn>
        </template>
        </v-layout>
    </v-container>
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
        },
        colorHash: {
            type: String,
            default: ""
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