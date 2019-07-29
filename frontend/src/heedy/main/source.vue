<template>
    <h1>{{ source.fullname }}</h1>
</template>
<script>
export default {
    props: {
        source: Object
    },
    beforeMount() {
        let typePath = this.$store.state.heedy.typePaths[this.source.type] || "";
        if (typePath!="") {
            this.$router.push(`/source/${this.source.id}/` + typePath);
        }
    }
}
</script>