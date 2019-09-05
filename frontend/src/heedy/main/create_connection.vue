<template>
    <v-content>
        <v-container>
            <v-layout column>
                <v-flex justify-center align-center text-center style="padding: 10px; padding-bottom: 20px;">
                            <h1 style="color:#1976d2;">Create a new Connection</h1>
                        
                </v-flex>
                <v-flex>
                <v-card>
                    <div style="padding: 10px; padding-bottom: 0;">
                        <v-alert v-if="alert.length>0" text outlined color="deep-orange" icon="error_outline">{{ alert }}</v-alert>
                    </div>
                    
                    <v-card-title>
                        <v-text-field label="Connection Name" placeholder="My Connection" v-model="name"></v-text-field>
                        <v-text-field label="Connection Name" placeholder="myconnection" v-model="name"></v-text-field>
                    </v-card-title>
                    <v-card-title>
                        <scope-editor v-model="scopes"></scope-editor>
                    </v-card-title>
                    <v-card-text class="text--primary" v-if="advanced">
                    <v-container fluid grid-list-md>
                    <v-layout row >
                        <v-flex sm5 md4 xs12>
                        <avatar-editor ref="avatarEditor" image="settings_input_component"></avatar-editor>
                        </v-flex>
                        <v-flex sm7 md8 xs12>
                        <v-container>
                        <v-textarea solo label="Description" v-model="description"></v-textarea>
                        </v-container>
                        </v-flex>
                    </v-layout>
                    </v-container>
                    
                    </v-card-text>
                    <v-card-actions>
                        <v-btn text @click="advanced = !advanced">
                            <v-icon left>{{advanced? "expand_less":"expand_more"}}</v-icon> Advanced
                        </v-btn>
                        <v-spacer>
                        </v-spacer>
                        <v-btn dark color="blue" @click="create" :loading="loading">
                            Create
                        </v-btn>
                    </v-card-actions>
                </v-card>
                </v-flex>
            </v-layout>
        </v-container>
    </v-content>
</template>
<script>

import {AvatarEditor, ScopeEditor} from "../components.mjs";


import api from "../api.mjs";
export default {
    components: {
        AvatarEditor,
        ScopeEditor
    },
    data: () => ({
        description: "",
        scopes: "",
        name: "",
        advanced: false,
        loading: false,
        alert: ""

    }),
    methods: {
        create: async function() {
            if (this.loading) return;

            this.loading = true;
            this.alert="";

            this.alert = "Unimplemented";
            console.log({
                description: this.description,
                scopes: this.scopes,
                name: this.name
            });
            this.loading = false;
        }
    }
}
</script>