<template>
    <div class='entry-content' v-for='(entry, idx) in entries' :key='idx'>
      <h2 :id='"e-" + entry.uniqueId'>{{entry.title}}</h2>
      <span v-html='entry.body'></span>
    </div>

</template>

<script>
  export default {
    props: {
      entryList: {
        type: Array
      }
    },
    data() {
      return {
        entries: this.entryList
      }
    }
  }
</script>
