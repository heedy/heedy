<template>
  <h-card-page :title="'Create ' + type.title" :alert="alert">
    <v-form @submit="create">
      <v-container fluid grid-list-md>
        <v-layout row>
          <v-flex sm5 md4 xs12>
            <h-icon-editor
              ref="iconEditor"
              :image="object.icon || ''"
              :defaultIcon="type.icon"
            ></h-icon-editor>
          </v-flex>
          <v-flex sm7 md8 xs12>
            <v-container>
              <v-text-field
                autofocus
                label="Name"
                :placeholder="`My ` + type.title"
                v-model="name"
              ></v-text-field>
              <v-text-field
                label="Description"
                placeholder="A short description goes here"
                v-model="description"
              ></v-text-field>
              <h-tag-editor v-model="tags" />
            </v-container>
          </v-flex>
        </v-layout>
        <slot
          :meta="object.meta"
          :on="{ input: (m) => $emit('input', { ...object, meta: m }) }"
        ></slot>
      </v-container>
      <v-container v-if="advanced && hasAdvanced">
        <slot
          name="advanced"
          :meta="object.meta"
          :on="{ input: (m) => $emit('input', { ...object, meta: m }) }"
        ></slot>
      </v-container>
      <v-card-actions>
        <v-btn v-if="hasAdvanced" text @click="advanced = !advanced">
          <v-icon left>{{ advanced ? "expand_less" : "expand_more" }}</v-icon
          >Advanced
        </v-btn>
        <v-spacer></v-spacer>
        <v-btn text @click="$router.go(-1)">Cancel</v-btn>
        <v-btn type="submit" color="primary" :loading="loading">Create</v-btn>
      </v-card-actions>
    </v-form>
  </h-card-page>
</template>
<script>
export default {
  model: {
    prop: "object",
    event: "input",
  },
  props: {
    object: Object,
    validator: {
      type: Function,
      default: (o) => "",
    },
  },
  data: () => ({ alert: "", advanced: false, loading: false }),
  computed: {
    hasAdvanced() {
      return !!this.$slots["advanced"];
    },
    type() {
      let otype = this.$store.state.heedy.object_types[this.object.type] || {};
      return {
        icon: "assignment",
        title: "Object",
        ...otype,
      };
    },
    name: {
      get() {
        return this.object.name || "";
      },
      set(v) {
        this.$emit("input", { ...this.object, name: v });
      },
    },
    description: {
      get() {
        return this.object.description || "";
      },
      set(v) {
        this.$emit("input", { ...this.object, description: v });
      },
    },
    tags: {
      get() {
        return this.object.tags || "";
      },
      set(v) {
        this.$emit("input", { ...this.object, tags: v });
      },
    },
  },
  methods: {
    create: async function (e) {
      e.preventDefault();
      if (this.loading) return;

      this.loading = true;
      this.alert = "";

      let obj = this.object;

      let vstring = this.validator(obj);
      if (vstring != "") {
        this.alert = vstring;
        this.loading = false;
        return;
      }

      if (this.$refs.iconEditor.hasImage()) {
        // We are in the image picker, and an image was chosen
        obj.icon = this.$refs.iconEditor.getImage();
      }

      // Attempt to
      let result = await this.$frontend.rest("POST", `api/objects`, obj);

      if (!result.response.ok) {
        this.alert = result.data.error_description;
        this.loading = false;
        return;
      }
      // The result comes without the icon, let's set it correctly
      if (obj.icon !== undefined) {
        result.data.icon = obj.icon;
      } else {
        result.data.icon = "";
      }

      this.$store.commit("setObject", result.data);
      this.loading = false;
      this.$router.replace({ path: `/objects/${result.data.id}` });
    },
  },
};
</script>