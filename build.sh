go build -o bin/colonels
# minify js (we'll probably need a real build system later)
# if this line fails for you, `go get github.com/tdewolff/minify/cmd/minify`
minify static/js/game.js -o static/js/game.min.js
minify static/js/game_setup.js -o static/js/game.min.js
./bin/colonels
