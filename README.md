## colonels.io

### what is this?
Last fall my friends and I played [generals.io](generals.io) for a while after finals and liked it a lot. I'm a bit of an optimization nerd (see my other repos), so I decided to build a clone in Go as a way of learning the language (with lots of help from https://gobyexample.com/, which was written by one of my old bosses!).

The main thing I had fun with was writing the protobuf bindings in JavaScript (turns out, if you're careful about synchronization you can send ~10 bytes over the wire), though it isn't practical for a decent number of reasons. I think websocket concurrency in Go is neat too.

In an ideal world I'd sit down and finish this, but there's a lot of neat research I want to work on instead, so this will have to wait. If someone wants to adopt it or learn from the codebase feel free.

### todo
It's unclear what's left to do - I'll need to dig into the code at some point. IIRC it's mostly stuff like
JavaScript game state management (i.e. has someone won, and how does the game server communicate that to the 
client). I'll dig into it later and see what it needs to finish.

### how to build

#### prerequisites
I'm currently on a bus without wifi, so some of these directions might be wrong, but I think you need:
1. A recent version of Go (I have go1.7.4 for Mac OSX (darwin)), the language used to write the server side 
code. You can either grab this from the [Go website](golang.org), or via Brew.
2. A recent version of `protoc` (I have libprotoc 3.1.0), which I think you can install via Brew. This is the 
tool that translates the protobuf spec for your messages (see /static/proto/main.proto for an example) into 
Go and JavaScript code that can encode/decode the binary messages sent over the wire (see main.pb.go for an 
example).

#### build
Just run:
``./build.sh``
You may need to `chmod +x build.sh` first, though that's unlikely. If you open the script, it:
1. Runs the protobuf translator, which generates the JavaScript/Go files necessary for message encode/decode
2. Runs the colonels.io codebase in debug mode using the Go interpreter.
