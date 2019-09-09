<template>
    <v-content>
        <v-container grid-list-xl>
                <v-layout fill-height column>
                    <v-flex>
                        <v-card>
                            <v-card-title>
                                
                                <v-list-item two-line style="margin-top: -8px;">
                                    <v-list-item-avatar>
                                        <avatar :image="connection.avatar" :colorHash="connection.id" />
                                    </v-list-item-avatar>
                                    <v-list-item-content>
                                        <v-list-item-title class="headline mb-1" >{{ connection.name }}</v-list-item-title>
                                        <v-list-item-subtitle>{{ connection.description }}</v-list-item-subtitle>
                                    </v-list-item-content>
                                    <v-list-item-action>
                                        <v-toolbar flat collapsed >
                                <v-toolbar-items style="margin-right: -30px;">
                                    <v-dialog v-model="showkey" width="500">
                                        <template #activator="{on: onDialog}">
                                            <v-tooltip bottom>
                                                <template #activator="{on}">
                                                    <v-btn icon v-on="{...on}" @click="() => {onDialog.click();getKey();}"><v-icon>vpn_key</v-icon></v-btn>
                                                </template>
                                                <span>View AccessToken</span>
                                            </v-tooltip>
                                        </template>
                                        <v-card>
                                            <v-card-title text-center>
                                                <v-layout row justify-center>
                                                <v-flex text-center style="padding-top: 15px; padding-bottom: 15px;">
                                                    <h3 style="color:#1976d2;padding-bottom: 7px;">Access Token</h3>
                                                    <h4>{{ token }}</h4>
                                                    <h6 style="color: gray;">You can reset the token in <v-icon>edit</v-icon></h6>
                                                    <v-btn rounded outlined style="margin-top: 15px;" color="grey" @click="showkey=false">Done</v-btn>
                                                </v-flex>
                                                </v-layout>
                                            </v-card-title>
                                        </v-card>
                                    </v-dialog>
                                    <v-tooltip bottom>
                                        <template #activator="{on}">
                                            <v-btn
                                            icon v-on="on"
                                            :to="`/connections/${connection.id}/update`"
                                            >
                                            <v-icon>edit</v-icon>
                                        </v-btn>
                                        </template>
                                        <span>Edit Connection</span>
                                    </v-tooltip>
                                    <v-tooltip bottom v-if="Object.keys(connection.setting_schema).length > 0">
                                        <template #activator="{on}">
                                            <v-btn 
                                                icon v-on="on"
                                                color="blue darken-2"
                                                :to="`/connections/${connection.id}/settings`"
                                                >
                                                <v-icon>fas fa-cog</v-icon>
                                            </v-btn>
                                        </template>
                                        <span>Connection Settings</span>
                                    </v-tooltip>

                                    
                                    
                                    
                                </v-toolbar-items>
                            </v-toolbar>
                                    </v-list-item-action>
                                </v-list-item>
                            </v-card-title>
                        </v-card>
                </v-flex>
            </v-layout>
        </v-container>
    </v-content>
</template>

<script>
import {Avatar} from "../components.mjs";
import api from "../api.mjs";
export default {
    components: {
        Avatar
    },
    data: () => ({
        showkey: false,
        token: "..."
    }),
    props: {
        connection: Object
    },
    watch: {
        showkey(newv) {
            this.token = "...";
        }
    },
    methods: {
        getKey: async function() {
            console.log("Reading access token for",this.connection.id);
            let result = await api(
                    "GET",
                    `api/heedy/v1/connections/${this.connection.id}`,
                    {token: true}
                );
            if (!result.response.ok) {
                this.token = result.data.error_description;
                return;
            }
            this.token = result.data.access_token;
        }
    }
}
</script>