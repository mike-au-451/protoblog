<template>
    <div class='posted-container'>
      <h2>Posted</h2>
      <ul>
        <li v-for='(entry, idx) in entries' :key='idx'>
          <a :href='"#e-" + entry.uniqueId'>{{entry.posted}}</a>
        </li>
      </ul>
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
    },
  }
</script>
