/* Use for setting up websockets and communication in the game layer */
var protocol = null;

define(['require', 'google-protobuf', 'zepto', 'main_pb', 'pixi'], function(require) {
	"use strict";
	var $ = require('zepto');
	var PIXI = require('pixi');

	var server = null;
	var url = "";
	var retrier = null;

	var meta = {ready: false};
	var CELL_SIZE = 30;

	// load the protobufs
	protocol = require('main_pb');

	$(document).ready(function() {		
		if (production) {
			url = "some production url";
		}
		else {
			url = "ws://localhost:8000/game/" + matchid + '/ws';
		}
		
		server = new WebSocket(url);
		server.binaryType = "arraybuffer";
		server.onmessage = onmessage;
		server.onopen = function() {
			$("#status").html('websocket connected');
		}
		server.onclose = function() {
			// should probably do exponential backoff
			$("#status").html('websocket closed. retrying in 3...');
			if (retrier == null) {
				retrier = setTimeout(function(){ 
					try {
						server = new WebSocket(url);
					}
					catch(e) {
						// fail gracefully
					}
					retrier = null;
				}, 3);			
			}
		}

		$("#start").click(function() {
			// send a ping to the backend to communicate
			if (server.readyState == WebSocket.OPEN) {
				if (meta.ready) {
					meta.ready = false;
					var msg = new proto.main.playerStatus();
					msg.setStatus(protocol.Status.UNREADY);
					server.send(msg.serializeBinary());
				}
				else {
					meta.ready = true;
					var msg = new proto.main.playerStatus();
					msg.setStatus(protocol.Status.READY);
					server.send(msg.serializeBinary());
				}
				startGame()			
			}
		});
	});

	/* when the server sends data, an update, etc. */
	function onmessage(event) {
		var data = JSON.parse(event.data);
		units.clear();
		for (var i = 0; i < data.Cells.length; i++) {
			for (var j = 0; j < data.Cells[i].length; j++) {
				if (data.Cells[i][j].CellType == 1) {
					units.beginFill(0xf44e42, 1);
					units.drawRect(i * CELL_SIZE, j * CELL_SIZE, CELL_SIZE, CELL_SIZE);
					units.endFill();
				}
			}
		}
	}

	function startGame() {
		window.renderer = PIXI.autoDetectRenderer(CELL_SIZE * 20, CELL_SIZE * 20);
		document.body.appendChild(renderer.view);

		window.stage = new PIXI.Container();	
		window.mainLayer = new PIXI.Container();
		// mainLayer.interactive = true;
		window.graphics = new PIXI.Graphics();
		window.units = new PIXI.Graphics();
		drawGrid(graphics, 20);
		mainLayer.addChild(graphics);
		mainLayer.addChild(units);

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
		graphics.beginFill(0xcee1ff, 0.5);
		graphics.drawRect(0, 0, CELL_SIZE * num_cells, CELL_SIZE * num_cells);
		graphics.endFill();
	}

	function setupTouch() {
		var old = {};
		var touchpoint = {};
		var mousedown = false;

		stage.interactive = true;

		var onDragStart = function(e) {
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

		var interaction = new PIXI.interaction.InteractionManager(renderer);

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
})