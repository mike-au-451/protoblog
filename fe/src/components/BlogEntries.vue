<template>
    <div class='entry-container'>
      <div class='entry-content' v-for='(entry, idx) in entries' :key='idx'>
        <h2 :id='"e-" + entry.uniqueId'>{{entry.title}}</h2>
        <span v-html='entry.body'></span>
      </div>
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

<style>
  .entry-container {
    flex-direction: column;
  }
</style>
