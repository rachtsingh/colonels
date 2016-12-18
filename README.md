## colonels.io

### todo
1. implement a `broadcast` function inside `game_state.go` so that we can send messages to the waiting client about how many people have joined, how many people are ready, etc. We'll need to put together a JSON/string based channel for this, maybe a struct withan enum and a string or something (not sure what the right Go idiom is)
  - actually it's not clear if I should just have one thread for each user, or whether it should be a read thread and a write thread. Basically, the issue is that readJSON isn't a channel like in select. It seems to block.
2. figure out the Pixi.js renderer and figure out how to push fast updates (shouldn't be too bad). - I'm working on this now
