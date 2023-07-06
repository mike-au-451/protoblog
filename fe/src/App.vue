<template>
  <header>
    <img src="./assets/cancelled.png" height="125" />

    <h1 class="wrapper">
      Its tradoffs...all the way down
    </h1>
  </header>

  <hr>
  
  <main>
    <PostedDates :entryList='entryList' />
    <BlogEntries :entryList='entryList' />
    <EntryTags :entryList='entryList' />
  </main>

  <hr>
  
  <footer>
  </footer>
</template>

<script>
  import PostedDates from './components/PostedDates.vue'
  import BlogEntries from './components/BlogEntries.vue'
  import EntryTags from './components/EntryTags.vue'

  export default {
    name: 'App',
    components: {
      PostedDates,
      BlogEntries,
      EntryTags,
    },
    data() {
      return {
        entryList: getEntryList()
      }
    }
  }

  function getEntryList() {
    var entryUrl = "http://localhost:7890/entries";
    var xmlHttp = new XMLHttpRequest();
    xmlHttp.open( "GET", entryUrl, false ); // false for synchronous request
    xmlHttp.send( null );

    var entries =  JSON.parse(xmlHttp.responseText);
    return entries;
  }

</script>

<style>
  #app {
    font-family: Avenir, Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    color: #2c3e50;
    margin-top: 10px;
  }

  header {
    display: flex;
    line-height: 1.5;
    column-gap: 20px;
  }

  main {
    display: flex;
    column-gap: 20px;
  }

  footer {
  }
</style>
