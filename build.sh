# build proto out
# both javascript (commonjs style) and golang
protoc --proto_path=static/proto --go_out=. --js_out=import_style=commonjs,binary:./static/js/ static/proto/main.proto
# ok, this is a sad day - we're going to need to patch the output from the 
# golang exporter in order to get requirejs working.
echo "define(function(require, exports, module){" | cat - static/js/main_pb.js > temp && mv temp static/js/main_pb.js
echo "});" >> static/js/main_pb.js

go build -o bin/colonels

# minify js (we'll probably need a real build system later)
# if this line fails for you, `go get github.com/tdewolff/minify/cmd/minify`
# minify static/js/game.js -o static/js/game.min.js
# minify static/js/game_setup.js -o static/js/game.min.js
./bin/colonels -debug
