<template>
    <div class='tag-container'>
      <span v-for='(entry) in entries' :key='entry.uniqueId'>
        <a v-for='(tag, idx) in entry.tags' :key='idx' :href='"#e-" + entry.uniqueId'>{{tag}}&nbsp;</a>
        <br>
      </span>
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
