## colonels.io

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
