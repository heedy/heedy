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
                    <v-container fluid grid-list-md>
                    <v-layout row >
                        <v-flex sm5 md4 xs12>
                        <avatar-editor ref="avatarEditor" image="settings_input_component"></avatar-editor>
                        </v-flex>
                        <v-flex sm7 md8 xs12>
                        <v-container>
                            <v-text-field label="Name" placeholder="My Connection" v-model="name"></v-text-field>
                            <v-text-field label="Description" placeholder="This connection does stuff" v-model="description"></v-text-field>
                            <scope-editor v-model="scopes"></scope-editor>
                        </v-container>
                        </v-flex>
                    </v-layout>
                    </v-container>
                    
                    <v-card-actions>
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
        loading: false,
        alert: ""

    }),
    methods: {
        create: async function() {
            if (this.loading) return;

            this.loading = true;
            this.alert="";

            let query = {
                name: this.name.trim(),
                description: this.description.trim(),
                scopes: this.scopes,
                avatar: this.$refs.avatarEditor.getImage()
            };

            if (query.name.length == 0) {
                this.alert = "A name is required"
                this.loading = false;
                return;
            }

            let result = await api(
                "POST",
                `api/heedy/v1/connections`,
                query
            );

            if (!result.response.ok) {
                this.alert = result.data.error_description;
                this.loading = false;
                return;
            }

            // The result comes without the avatar, let's set it correctly
            result.data.avatar = query.avatar;

            this.$store.commit("setConnection",result.data);
            this.loading = false;
            this.$router.push({path: `/connections/${result.data.id}`});
        }
    }
}
</script>