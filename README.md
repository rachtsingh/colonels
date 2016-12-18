## colonels.io

### todo
1. implement a `broadcast` function inside `game_state.go` so that we can send messages to the waiting client about how many people have joined, how many people are ready, etc. We'll need to put together a JSON/string based channel for this, maybe a struct withan enum and a string or something (not sure what the right Go idiom is)

2. figure out the Pixi.js renderer and figure out how to push fast updates (shouldn't be too bad). - I'm working on this now
