Notes from manual pass 
- The AI can do a lot of boilerplate code quickly but the subtleties are beyond it. Nothing compares to the finesse of a good coder.
- Actor spawing message is sigltly broken. the type of the message is wrong.


- Router right now has a map name:actor. What if want to have multiple actors working?
- Round-Robin with a name:[list of actors] can be a possibility.

- in the future we may want to hash or cryprographically secure the code for an actor
- so a bad guy can't overwrite our actors so easily.

- a good feature would be to have some security embedded into the language by default
as relay is distributed and etc...

- a nice auto UI would be very nice for this, we want easyness of operation anyway.

- I dont even want to see how websocket actor is done. its terrifying
- same but for federated one.

- right now with the syntax that we have we cant really generate servers dynamically.
- furthermore, server is seems to not be working on the repl or when evaluating..
