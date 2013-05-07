twittergo-examples
==================
These examples should give a good idea of how to use the
https://github.com/kurrik/twittergo client library.

Using
-----
First install the following dependencies:

    go get -u github.com/kurrik/twittergo
    go get -u github.com/kurrik/oauth1a

Then add a file called
`CREDENTIALS` in this project root.  The format of this file should be:

    <Twitter consumer key>
    <Twitter consumer secret>
    <Twitter access token>
    <Twitter access token secret>

Note that some examples (like `tweet`) actually write to
the API, so use a testing account!

To run an example:

    go run <path to example>/main.go

The simplest example is probably `verify_credentials`.  This calls an
endpoint which will return the current user if the request is signed
correctly.

App Engine
----------
The Google App Engine examples are a bit more involved, mostly because
you need to bundle a copy of the library with your app.  To facilitate this
I've chosen to utilize git submodules.  After checking out this repo, run:

    git submodule init
    git submodule update

You may need to run git submodule update from time to time as I update the
example to use a more current version of the library.

There are some dependencies you'll also need to satisfy. The following only
need to be done once per machine:

    brew install pkill
    sudo npm install -g grunt-cli
    <Install Go dev appserver to ~/src/google_appengine_go>

Per-project:

    cd <PROJECT_DIR>
    npm install
    grunt develop

Examples will be accessible on `http://localhost:9996`.

