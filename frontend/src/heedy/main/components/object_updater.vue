<template>
  <h-card-page :title="'Edit ' + object.name" :alert="alert">
    <v-container fluid grid-list-md>
      <v-layout row>
        <v-flex sm5 md4 xs12>
          <h-icon-editor
            ref="iconEditor"
            :image="object.icon"
            :colorHash="object.id"
            :defaultIcon="type.icon"
          ></h-icon-editor>
        </v-flex>
        <v-flex sm7 md8 xs12>
          <v-container>
            <v-text-field autofocus label="Name" :placeholder="`My ` + type.title" v-model="name"></v-text-field>
            <v-text-field
              label="Description"
              placeholder="A short description goes here"
              v-model="description"
            ></v-text-field>
            <h-tag-editor v-model="tags" />
          </v-container>
        </v-flex>
      </v-layout>
      <slot></slot>
    </v-container>
    <v-container v-if="advanced && hasAdvanced">
      <slot name="advanced"></slot>
    </v-container>
    <v-card-actions>
      <v-btn dark color="red" @click="del" :loading="loading">Delete</v-btn>
      <v-btn v-if="hasAdvanced" text @click="advanced = !advanced">
        <v-icon left>{{ advanced ? "expand_less" : "expand_more" }}</v-icon>Advanced
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn dark color="blue" @click="update" :loading="loading">Save</v-btn>
    </v-card-actions>
  </h-card-page>
</template>
<script>
export default {
  props: {
    object: Object,
    meta: {
      type: Object,
      default: () => ({})
    }
  },
  data: () => ({
    alert: "",
    modified: {},
    advanced: false,
    loading: false
  }),
  methods: {
    update: async function() {
      if (this.loading) return;

      this.loading = true;
      this.alert = "";

      let modified = { ...this.modified };

      if (this.$refs.iconEditor.hasImage()) {
        // We are in the image picker, and an image was chosen
        modified.icon = this.$refs.iconEditor.getImage();
      }

      if (Object.keys(this.meta).length > 0) {
        modified.meta = this.meta;
      }

      if (Object.keys(modified).length > 0) {
        console.log("UPDATING", modified);
        let result = await this.$frontend.rest(
          "PATCH",
          `api/objects/${this.object.id}`,
          modified
        );

        if (!result.response.ok) {
          this.alert = result.data.error_description;
          this.loading = false;
          return;
        }
        this.$store.dispatch("readObject", {
          id: this.object.id
        });
      }

      this.loading = false;
      this.$router.push(`/objects/${this.object.id}`);
    },
    del: async function() {
      let s = this.object;
      if (
        confirm(
          `Are you sure you want to delete '${this.object.name}'? This deletes all associated data.`
        )
      ) {
        let res = await this.$frontend.rest(
          "DELETE",
          `/api/objects/${this.object.id}`
        );
        if (!res.response.ok) {
          this.alert = res.data.error_description;
        } else {
          this.alert = "";
          if (s.app != null) {
            this.$router.push(`/apps/${s.app}`);
          } else {
            this.$router.push(`/users/${s.owner}`);
          }
        }
      }
    }
  },
  computed: {
    description: {
      get() {
        if (this.modified.description !== undefined) {
          return this.modified.description;
        }
        return this.object.description;
      },
      set(v) {
        this.$frontend.vue.set(this.modified, "description", v);
      }
    },
    name: {
      get() {
        if (this.modified.name !== undefined) {
          return this.modified.name;
        }
        return this.object.name;
      },
      set(v) {
        this.$frontend.vue.set(this.modified, "name", v);
      }
    },
    tags: {
      get() {
        if (this.modified.tags !== undefined) {
          return this.modified.tags;
        }
        return this.object.tags;
      },
      set(v) {
        this.$frontend.vue.set(this.modified, "tags", v);
      }
    },

    type() {
      let otype = this.$store.state.heedy.object_types[this.object.type] || {};
      return {
        icon: "assignment",
        title: "Object",
        ...otype
      };
    },
    hasAdvanced() {
      return !!this.$slots["advanced"];
    }
  }
};
</script>