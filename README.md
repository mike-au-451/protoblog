# Protoblog

The blog from the primordial ooze.
This is a proof of concept and a mechanism for dogfooding my software.  At the moment the workflow is too manual, and the web site crashes as soon as you close the remote terminal.

## TODO

1.  ☑ Fix or replace zerolog.  The time format is broken (which appears to be a known bug).
2.  ☐ Open log files so the blog can run without a terminal.
3.  ☐ Clean up the workflow so you dont need to ssh into the server to complete a blog upload.
4.  ☐ Use rsync rather than scp.
5.  ☐ Allow links between blog entries.
6.  ☐ Implement tags.

001-server-render

`markdown-it` is a server-side rendering engine which uses Nodejs specific features.
Attempting to use it client-side proved difficult, so move back to server-side rendering
using the `goldmark` package.

