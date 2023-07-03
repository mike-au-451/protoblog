<template>
  <ul>
    <li v-for='(entry, idx) in entries' :key='idx'>
      <a :href='"#e-" + entry.uniqueId'>{{entry.posted}}</a>
    </li>
  </ul>
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
    },
  }

</script>
