<template >
  <v-content>
    <v-container fill-height>
      <v-layout>
        <v-flex>
          <v-card>
            <v-container>
              <slot></slot>
            </v-container>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </v-content>
</template>
