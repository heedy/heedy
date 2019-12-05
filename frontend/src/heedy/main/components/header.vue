<template>
  <v-flex>
    <v-card>
      <v-card-title>
        <v-list-item
          v-if="showTitle"
          two-line
          style="margin-top: -10px;margin-bottom: -8px; overflow: hidden;"
        >
          <v-list-item-avatar>
            <h-icon :image="icon" :colorHash="colorHash" />
          </v-list-item-avatar>
          <v-list-item-content>
            <v-list-item-title class="headline mb-1">{{ name }}</v-list-item-title>
            <v-list-item-subtitle>{{ description }}</v-list-item-subtitle>
          </v-list-item-content>
          <v-list-item-action style="overflow: hidden">
            <v-toolbar flat collapsed>
              <v-toolbar-items style="margin-right: -25px;">
                <slot></slot>
              </v-toolbar-items>
            </v-toolbar>
          </v-list-item-action>
        </v-list-item>
        <v-toolbar v-else flat collapsed>
          <div v-if="showIcon" style="padding-right: 20px">
            <h-icon :image="icon" :colorHash="colorHash" />
          </div>
          <v-spacer></v-spacer>
          <v-toolbar-items style="align-items: center">
            <slot></slot>
          </v-toolbar-items>
        </v-toolbar>
      </v-card-title>
    </v-card>
  </v-flex>
</template>
<script>
export default {
  props: {
    icon: String,
    colorHash: {
      type: String,
      default: ""
    },
    name: String,
    description: String,
    showTitle: {
      type: Boolean,
      default: true
    },
    showIcon: {
      type: Boolean,
      default: true
    }
  }
};
</script>