# Ace Away!

[![Build Status](https://travis-ci.org/wafflespeanut/ace-away.svg?branch=master)](https://travis-ci.org/wafflespeanut/ace-away)

"Ace" is a card game that I play with my friends every now and then. It's somewhat popular in India, but there are [very few references](https://boardgames.stackexchange.com/q/7902/) addressing this game. This is a hobby project which implements that game as a web app using Go and Vue.

[Live demo](https://waffles.space/ace-away).

### Usage

> **NOTE:** This requires `npm` and `go ^1.11`.

```
make prepare
make serve
```

- Visit `localhost:3000` in your browser.
- Create a room with some set number of players.
- Your friends can now join that room (as long as they're part of the same network).
- The game will begin once the room has enough players.
