/* Use for setting up websockets and communication in the game layer */
var server = null;
var meta = {ready: false};
var CELL_SIZE = 30;

$(document).ready(function() {
	// set up a websocket connection
	if (production) {
		server = new WebSocket("some production url");
	}
	else {
		server = new WebSocket("ws://localhost:8000/game/" + matchid + '/ws');
	}
	server.onmessage = onmessage;

	$("#start").click(function() {
		// send a ping to the backend to communicate
		if (meta.ready) {
			meta.ready = false;
			server.send(JSON.stringify({m: 'unready'}));
		}
		else {
			meta.ready = true;
			server.send(JSON.stringify({m: 'ready'}));
		}

		startGame()
	});
})

/* when the server sends data, an update, etc. */
function onmessage(event) {
	var data = JSON.parse(event.data);
	console.log(event);
}

function startGame() {
	renderer = PIXI.autoDetectRenderer(CELL_SIZE * 20, CELL_SIZE * 20);
	document.body.appendChild(renderer.view);

	stage = new PIXI.Container();	
	mainLayer = new PIXI.Container();
	// mainLayer.interactive = true;
	graphics = new PIXI.Graphics();
	drawGrid(graphics, 20);
	mainLayer.addChild(graphics);

	stage.addChild(mainLayer);
	renderer.render(stage);
	requestAnimationFrame(animate);

	setupTouch();
}

function animate() {
	renderer.render(stage);
	requestAnimationFrame(animate);
}

function drawGrid (graphics, num_cells) {
	graphics.lineStyle(1, 0x9ec3ff, 0.7);
	// vertical lines
	for (var i = 0; i < num_cells + 1; i++) {
		graphics.moveTo(i * CELL_SIZE, 0);
		graphics.lineTo(i * CELL_SIZE, CELL_SIZE * num_cells);
	}
	// horizontal lines
	for (var i = 0; i < num_cells + 1; i++) {
		graphics.moveTo(0, i * CELL_SIZE);
		graphics.lineTo(CELL_SIZE * num_cells, i * CELL_SIZE);
	}
}

function setupTouch() {
	var old = {};
	var touchpoint = {};
	var mousedown = false;

	stage.interactive = true;

	var onDragStart = function(e) {
		console.log("started!");
		old.x = mainLayer.position.x;
		old.y = mainLayer.position.y;
		touchpoint.x = e.data.global.x;
		touchpoint.y = e.data.global.y;
		mousedown = true;
	}

	var onDragEnd = function(e) {
		mousedown = false;
	}

	var onDragMove = function(e) {
		if (mousedown) {
			mainLayer.position.x = old.x - (touchpoint.x - e.data.global.x);
			mainLayer.position.y = old.y - (touchpoint.y - e.data.global.y);
		}
	}

	interaction = new PIXI.interaction.InteractionManager(renderer);

	// interaction.on('mousedown', onDragStart);
	// interaction.on('mouseup', onDragEnd);
	// interaction.on('mousemove', onDragMove);
	// stage.interactionManager.onMouseUp = onDragEnd;
	interaction
        // events for drag start
        .on('mousedown', onDragStart)
        .on('touchstart', onDragStart)
        // events for drag end
        .on('mouseup', onDragEnd)
        .on('mouseupoutside', onDragEnd)
        .on('touchend', onDragEnd)
        .on('touchendoutside', onDragEnd)
        // events for drag move
        .on('mousemove', onDragMove)
        .on('touchmove', onDragMove);
}