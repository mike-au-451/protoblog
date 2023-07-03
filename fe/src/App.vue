<template>
  <div class='header'>
    <img src='./assets/cancelled.png' height='100' />
    <h1>Its Tradeoffs, All The Way Down...</h1>
  </div>
  <hr>

  <div class='content'>
    <div id='posted'>
      <h2>Posted</h2>
      <posted-date :entryList='entryList'></posted-date>
    </div>
    <div id='content'>
      <h2>Content</h2>
      <blog-entry :entryList='entryList'></blog-entry>
    </div>
    <div id='tags'>
      <h2>Tags</h2>
      <entry-tag :entryList='entryList'></entry-tag>
    </div>
  </div>
  <hr>

  <div class='footer'>
    <p>Blah blah...
    </p>
  </div>
</template>

<script>
  import PostedDate from "./components/PostedDate.vue";
  import BlogEntry from "./components/BlogEntry.vue";
  import EntryTag from "./components/EntryTag.vue";

  // TODO:
  // This seems like it should be loaded just once, and used as needed.
  var MarkdownIt = require('markdown-it');
  var md = new MarkdownIt();
  md.use(
      require('markdown-it-container'),
      'meta',
      {
          render: function(tokens, idx) {
              if (tokens[idx].nesting === 1) {
                  return "<!--\n";
              }
              else {
                  return "-->\n";
              }
          }
      }
  );

  export default {
    name: 'App',
    components: {
      PostedDate,
      BlogEntry,
      EntryTag,
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
      for (var ii = 0; ii < entries.length; ii++) {
        entries[ii].body = md.render(entries[ii].body)
      }
      // console.log(entries);
      return entries;
  }
</script>

<style>
  #app {
    font-family: Avenir, Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    color: #2c3e50;
    margin-top: 60px;
  }

  .header {
    display: flex;
    column-gap: 20px;
  }
  .content {
    display: flex;
    column-gap: 20px;
  }
  .footer {
    font-size: x-small;
  }

  #posted {
  }
  #content {
    flex-direction: column;
    flex: 1 0 70%;
  }
  #tags {
    flex: 1 0 10%;
  }
</style>
