/* Use for setting up websockets and communication in the game layer */
var server = null;
var meta = {ready: false};

$(document).ready(function()) {
	// set up a websocket connection
	if (production) {
		server = new WebSocket("some production url");
	}
	else {
		server = new WebSocket("ws://localhost:8000/game/" + matchid);
	}
	server.onmessage = onmessage;
	server.send("c");
}

/* when the server sends data, an update, etc. */
function onmessage(event) {
	var data = JSON.parse(event.data);
	if (data.type == "u") {
		
	}
	else if (data.type == "m") {
	}
}

$("#start").click(function() {
	// send a ping to the backend to communicate
	if (meta.ready) {
		meta.ready = false;
		server.send("d");
	}
	else {
		meta.ready = true;
		server.send("r");
	}
})
