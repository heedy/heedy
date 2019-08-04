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
                <hot-table :data="datatable" :rowHeaders="true" :colHeaders="true"></hot-table>
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
    data: () => ({
      datatable: [
          ["", "Ford", "Volvo", "Toyota", "Honda"],
          ["2016", 10, 11, 12, 13],
          ["2017", 20, 11, 14, 13],
          ["2018", 30, 15, 12, 13]
        ]
    }),
    props: {
        source: Object
    },
};
</script>