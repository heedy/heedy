<template>
    <div>
        <v-content>
            <v-container grid-list-xl>
                <v-layout fill-height column>
                    <v-flex>
                        <v-card>
                            <v-card-title>
                                <v-list-item two-line>
                                    <v-btn
                                    color="blue darken-2"
                                    dark
                                    fab
                                    right
                                    absolute
                                    to="/create/connection"
                                    >
                                    <v-icon>add</v-icon>
                                    </v-btn>
                                    <v-list-item-content>
                                        <v-list-item-title class="headline mb-1">Connections</v-list-item-title>
                                        <v-list-item-subtitle>External services and apps connected to heedy</v-list-item-subtitle>
                                    </v-list-item-content>
                                </v-list-item>
                                
                            </v-card-title>
                            <v-container>
                            <v-row no-gutters>
                            <template v-for="n in 4">
                                <v-col :key="n" sm="12" md="6" lg="4">
                                <v-card
                                    class="pa-2"
                                    outlined
                                    tile
                                >
                                    <v-list-item two-line subheader :to="`/connections/myid`">
                                    <v-list-item-avatar>
                                        <avatar image="person" colorHash="2" ></avatar>
                                    </v-list-item-avatar>
                                    <v-list-item-content>
                                        <v-list-item-title>Fitbit</v-list-item-title>
                                        <v-list-item-subtitle>Connection to the Fitbit Servers</v-list-item-subtitle>
                                    </v-list-item-content>
                                    </v-list-item>
                                </v-card>
                                </v-col>
                                
                            </template>

                            </v-row>
                        </v-container>
                        </v-card>
                    </v-flex>
                </v-layout>
                
            </v-container>
        </v-content>
    </div>
</template>
<script>
import {Loading, Avatar} from "../components.mjs";
export default {
    components: {
    Loading,
    Avatar
  }
}
</script>