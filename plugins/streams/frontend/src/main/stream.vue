<template>
<v-content>
      <v-container grid-list-xl>
        <v-layout fill-height column>
          <v-flex>
            <stream-info :stream="source" :editable="true"></stream-info>
          </v-flex>
          <v-flex>
            <v-card>
              <v-card-text>
                
              </v-card-text>
            </v-card>
          </v-flex>
          
        </v-layout>
        
      </v-container>
    </v-content>
</template>
<script>

import StreamInfo from "./streaminfo.vue";

export default {
    components: {StreamInfo},
    
    props: {
        source: Object
    },
};
</script>