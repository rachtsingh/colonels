/* Use for setting up websockets and communication in the game layer */
define(['require', 'google-protobuf', 'zepto', 'main_pb', 'pixi'], function(require) {
	"use strict";
	var $ = require('zepto');
	var PIXI = require('pixi');

	var server = null;
	var url = "";
	var retrier = null;

	var meta = {ready: false};
	var CELL_SIZE = 50;

	var playerColors = [
		0x000000,
		0xc0392b,
		0x2980b9,
		0x2ecc71,
		0x2c3e50,
		0x8e44ad,
		0xe67e22,
		0x16a085,
		0x7f8c8d,
		0x795548,
	]; // what to do if there are more than 9 colors? 
	// we can start adding borders around them.

	var gameReady = false;

	var MOVEID = 1;

	// load the protobufs
	window.protocol = require('main_pb');

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
				var msg = new protocol.clientToServer();
				msg.setWhich(protocol.clientMessageType.CLIENTSTATUS);
				if (meta.ready) {
					meta.ready = false;
					msg.setStatus(protocol.clientStatus.UNREADY);
				}
				else {
					meta.ready = true;
					msg.setStatus(protocol.clientStatus.READY);
				}
				server.send(msg.serializeBinary());
				startGame();
			}
		});
	});

	/* when the server sends data, an update, etc. */
	function onmessage(event) {
		var msg = protocol.ServerToClient.deserializeBinary(event.data).toObject();
		switch (msg.which) {
			case protocol.serverMessageType.FULLBOARD:
				// do some decoding here
				for (var i = 0; i < msg.board.rowsList.length; i++) {
					var column = msg.board.rowsList[i].columnList;
					for (var j = 0; j < column.length; j++) {
						// do we want to clone or just copy ptrs here?
						window.board[i][j] = Object.assign({}, column[j]);
					}
				}
				break;
			case protocol.serverMessageType.SINGLECELLUPDATE:
				window.board[msg.update.x][msg.update.y] = Object.assign({}, msg.update.value);
				break;
		}
		if (gameReady) {
			drawBoard();
		}
	}

	function startGame() {
		initializeBoard();
		// this is kind of ugly but I'll figure it out eventually
		// PIXI.settings.SCALE_MODE = PIXI.SCALE_MODES.NEAREST;
		PIXI.settings.RESOLUTION = window.devicePixelRatio;
		PIXI.loader
		.add([
			"../static/img/mountain.png",
			"../static/img/crown.png",
		])
		.load(function() {
			window.renderer = new PIXI.CanvasRenderer(
				$(window).width()/window.devicePixelRatio - 50, 
				$(window).height()/window.devicePixelRatio - 50, 
				{resolution: window.devicePixelRatio, antialias: true}
			);
			document.body.appendChild(renderer.view);
			
			// this is the main container (don't move it!)
			// it'll contain all of the HUD type update stuff
			window.stage = new PIXI.Container();	

			// this the layer with the game content on it. It moves with the mouse
			window.mainLayer = new PIXI.Container();

			// this the background of the board, which is static so is cached as a bitmap
			window.graphics = new PIXI.Graphics();
			graphics.cacheAsBitmap = true;
			drawGrid(graphics, 20);
			mainLayer.addChild(graphics);

			// this is a dynamically refreshing graphics layer showing territory
			window.territory = new PIXI.Graphics();
			mainLayer.addChild(territory);
			// this is where all the dynamic, moving sprites will live. On top of the graphics!
			window.units = new PIXI.Container();
			mainLayer.addChild(units);

			// numeric indicators
			window.textIndicators = [];
			initializeText()

			stage.addChild(mainLayer);
			renderer.render(stage);

			gameReady = true;
			requestAnimationFrame(animate);

			// set up game time variables?
			setupTouch();
		});
	}

	function initializeBoard() {
		window.board = new Array(20);
		for (var i = 0; i < 20; i++) {
			board[i] = new Array(20);
			for (var j = 0; j < 20; j++) {
				board[i][j] = {
					troops: 0,
					owner: 0,
					type: 0,
				};
			}
		}
	}

	function initializeText() {
		for (var i = 0; i < board.length; i++) {
			textIndicators.push([]);
			for (var j = 0; j < board[i].length; j++) {
				textIndicators[i].push(new PIXI.Text('0', {fontFamily : 'Arial', fontSize: 18, fill : 0xffffff, align : 'center'}));
				textIndicators[i][j].position.x = CELL_SIZE * (i + 0.5);
				textIndicators[i][j].position.y = CELL_SIZE * (j + 0.5);
				textIndicators[i][j].anchor.x = 0.5;
				textIndicators[i][j].anchor.y = 0.5;
			}
		}
	}

	function drawBoard() {
		units.removeChildren();
		territory.clear();
		for (var i = 0; i < board.length; i++) {
			for (var j = 0; j < board[i].length; j++) {
				var unit = null;
				if (board[i][j].type == protocol.squareType.MOUNTAIN) {
					unit = new PIXI.Sprite(
						PIXI.loader.resources["../static/img/mountain.png"].texture
					);
				}
				else if (board[i][j].type == protocol.squareType.CAPITAL) {
					unit = new PIXI.Sprite(
						PIXI.loader.resources["../static/img/crown.png"].texture
					);
				}
				if (unit != null) {
					units.addChild(unit);
					unit.anchor.set(0.5, 0.5);
					unit.position.x = (i + 0.5) * CELL_SIZE;
					unit.position.y = (j + 0.5) * CELL_SIZE;
					unit.width = CELL_SIZE;
					unit.height = CELL_SIZE;
				}
				// territory goes on bottom of units
				if (board[i][j].owner != 0) {
					territory.beginFill(playerColors[board[i][j].owner], 0.5)
					territory.drawRect(i * CELL_SIZE, j * CELL_SIZE, CELL_SIZE, CELL_SIZE);
					territory.endFill();
				}
				updateNumbers(i, j);
			}
		}
	}

	function updateNumbers(i, j) {
		if (board[i][j].owner == 0) {
			units.removeChild(textIndicators[i][j]);
		}
		else {
			units.removeChild(textIndicators[i][j]);
			textIndicators[i][j].text = board[i][j].troops.toString();
			units.addChild(textIndicators[i][j]);
		}
	}

	function animate() {
		renderer.render(stage);
		requestAnimationFrame(animate);
	}

	function drawBackground(parent) {
		var tilingSprite = PIXI.extras.TilingSprite.fromImage('/static/img/footer_lodyas.png', 1200, 1200);
		parent.addChild(tilingSprite);
		tilingSprite.position.x = -300;
		tilingSprite.position.y = -300;
	}

	function drawGrid (graphics, num_cells) {
		graphics.lineStyle(1, 0, 0);
		graphics.beginFill(0xffffff, 1);
		graphics.drawRect(0, 0, CELL_SIZE * num_cells, CELL_SIZE * num_cells);
		graphics.endFill();
		graphics.lineStyle(2, 0x7f8c8d, 1);
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
		graphics.lineStyle(1, 0, 0);
	}

	function setupTouch() {
		var old = {};
		var touchpoint = {};
		var mousedown = false;

		stage.interactive = true;

		var selectionIndicator = new PIXI.Graphics();
		selectionIndicator.lineStyle(3, 0x9b59b6, 1);
		selectionIndicator.drawRect(0, 0, CELL_SIZE, CELL_SIZE);

		// create the arrow to indicate you can move
		var pointerArrow = new PIXI.Graphics();
		pointerArrow.beginFill(0x000, 0.2);
		pointerArrow.drawRect(0, -1*CELL_SIZE/4, CELL_SIZE, CELL_SIZE);
		pointerArrow.endFill();
		pointerArrow.beginFill(0x9b59b6, 1);
		pointerArrow.drawPolygon([0, 0, CELL_SIZE/3, CELL_SIZE/4, 0, CELL_SIZE/2]);
		pointerArrow.endFill();
		var hovering = false;

		var currentDrag = {
			moved: false
		}

		var onDragStart = function(e) {
			old.x = mainLayer.position.x;
			old.y = mainLayer.position.y;
			touchpoint.x = e.data.global.x;
			touchpoint.y = e.data.global.y;
			mousedown = true;

			// reset whether it's been moved
			currentDrag.moved = false;
		}

		var onDragEnd = function(e) {
			mousedown = false;
			var newcellx = Math.floor((e.data.global.x - mainLayer.position.x)/CELL_SIZE);
			var newcelly = Math.floor((e.data.global.y - mainLayer.position.y)/CELL_SIZE);

			if (hovering) {
				// send a move up to the client
				pushCommand(selectionIndicator.x/CELL_SIZE, selectionIndicator.y/CELL_SIZE, newcellx, newcelly);
			}
			
			if (!currentDrag.moved) {
				mainLayer.removeChild(selectionIndicator);
				mainLayer.addChild(selectionIndicator);
				if (isValid(newcellx, newcelly)) {
					// these are relative to the mainLayer positions
					selectionIndicator.position.x = newcellx * CELL_SIZE;
					selectionIndicator.position.y = newcelly * CELL_SIZE;				
				}
			}

			checkHover(e);
		}

		// just check that the position isn't beyond board boundaries or a mountain
		var isValid = function(newcellx, newcelly) {
			if (newcellx < 0 || newcellx >= 20 || newcelly < 0 || newcelly >= 20) {
				return false;
			}
			else if (window.board[newcellx][newcelly].type == protocol.squareType.MOUNTAIN) {
				return false;
			}
			else {
				return true;
			}
		}

		// check whether the mouse position is next to the selected position
		var checkHover = function(e) {
			var newcellx = Math.floor((e.data.global.x - mainLayer.position.x)/CELL_SIZE);
			var newcelly = Math.floor((e.data.global.y - mainLayer.position.y)/CELL_SIZE);
			
			// set it to true, and if not just turn it off
			hovering = true;
			if ((newcellx - selectionIndicator.x/CELL_SIZE) == 1 && (newcelly - selectionIndicator.y/CELL_SIZE) == 0 && isValid(newcellx, newcelly)) {
				selectionIndicator.addChild(pointerArrow);
				pointerArrow.position.x = CELL_SIZE;
				pointerArrow.position.y = CELL_SIZE/4;
			}
			else if ((newcellx - selectionIndicator.x/CELL_SIZE) == -1 && (newcelly - selectionIndicator.y/CELL_SIZE) == 0 && isValid(newcellx, newcelly)) {
				selectionIndicator.addChild(pointerArrow);
				pointerArrow.rotation = Math.PI;
				pointerArrow.position.y = 3*CELL_SIZE/4;					
			}
			else if ((newcellx - selectionIndicator.x/CELL_SIZE) == 0 && (newcelly - selectionIndicator.y/CELL_SIZE) == 1 && isValid(newcellx, newcelly)) {
				selectionIndicator.addChild(pointerArrow);
				pointerArrow.rotation =	Math.PI/2;
				pointerArrow.x = 3*CELL_SIZE/4;
				pointerArrow.y = CELL_SIZE;
			}
			else if ((newcellx - selectionIndicator.x/CELL_SIZE) == 0 && (newcelly - selectionIndicator.y/CELL_SIZE) == -1 && isValid(newcellx, newcelly)) {
				selectionIndicator.addChild(pointerArrow);
				pointerArrow.rotation =	3*Math.PI/2;
				pointerArrow.x = CELL_SIZE/4;
				pointerArrow.y = 0;					
			}
			else {
				selectionIndicator.removeChild(pointerArrow);
				pointerArrow.position.x = 0;
				pointerArrow.position.y = 0;
				pointerArrow.rotation = 0;
				hovering = false;
			}
		}

		var onDragMove = function(e) {
			if (mousedown) {
				var newx = old.x - (touchpoint.x - e.data.global.x);
				var newy = old.y - (touchpoint.y - e.data.global.y)
				if (Math.abs(newx - old.x) > 10 || Math.abs(newy - old.y) > 10) {
					mainLayer.position.x = newx;
					mainLayer.position.y = newy;
					currentDrag.moved = true;
				}
			}
			else {
				checkHover(e);
			}
			// prevent moving too far
			if (mainLayer.position.x > 300) {
				mainLayer.position.x = 300;
			}
			if (mainLayer.position.x < -300) {
				mainLayer.position.x = -300;
			}
			if (mainLayer.position.y > 300) {
				mainLayer.position.y = 300;
			}
			if (mainLayer.position.y < -600) {
				mainLayer.position.y = -600;
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

	function clearQueue() {
		var msg = new protocol.clientToServer();
		msg.setWhich(protocol.clientMessageType.CANCELQUEUE);
		var cancel = new protocol.cancelQueue();
		cancel.setId(MOVEID);
		msg.setCancel(cancel);
		server.send(msg.serializeBinary());

		MOVEID = MOVEID + 1;
	}

	// send a move from (a, b) to (x, y)
	function pushCommand(a, b, x, y) {
		var msg = new protocol.clientToServer();
		msg.setWhich(protocol.clientMessageType.PLAYERMOVEMENT);
		var movement = new protocol.playerMovement();
		movement.setOldx(a);
		movement.setOldy(b);
		movement.setNewx(x);
		movement.setNewy(y);
		movement.setId(MOVEID);
		msg.setMovement(movement);
		server.send(msg.serializeBinary());

		MOVEID = MOVEID + 1;
	}
})