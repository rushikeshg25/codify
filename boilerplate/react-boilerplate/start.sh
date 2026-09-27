#!/bin/sh
#
# Starts the two processes a codeground needs: the user's dev server, and the
# agent that serves the terminal and file APIs.
#
# This script used to start only the agent and then block on `tail -f /dev/null`
# — a stray `&&` meant the dev server line was a no-op — so the preview URL
# served nothing until the user worked out that they had to run `npm run dev`
# themselves.

cd /app/codeground && npm run dev &

cd /app/wss && node index.js &
AGENT_PID=$!

# The container lives as long as the agent does, not the dev server. Stopping
# and restarting the dev server from the terminal is a normal thing to do and
# must not kill the pod out from under the person doing it.
wait "$AGENT_PID"
