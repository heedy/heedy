<template >
  <v-content>
    <v-container fill-height>
      <v-layout>
        <v-flex>
          <v-card>
            <v-card-text class="text--primary">
              <v-container>
                <slot></slot>
              </v-container>
            </v-card-text>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </v-content>
</template>
