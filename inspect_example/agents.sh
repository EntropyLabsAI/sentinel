#!/bin/bash

# Name of the tmux session
SESSION_NAME="my_session"

# Randomly set APPROVAL_YAML to approval_n.yaml or approval.yaml with 50% probability
if (( RANDOM % 2 )); then
    APPROVAL_YAML="approval.yaml"
else
    APPROVAL_YAML="approval_n.yaml"
fi

# Read tasks into an array

# Number of tasks
PANE_COUNT=4

# Start a new tmux session in detached mode
tmux new-session -d -s $SESSION_NAME

# Split the window into panes
for ((i=1; i<$PANE_COUNT; i++))
do
    tmux split-window -t $SESSION_NAME
    tmux select-layout -t $SESSION_NAME tiled > /dev/null
done

# Send the command to run the Go program with each task
for ((i=0; i<$PANE_COUNT; i++))
do
    TASK="${TASKS[$i]}"
    # Escape double quotes in the task string
    ESCAPED_TASK=$(printf '%q' "$TASK")
    tmux send-keys -t $SESSION_NAME.$i "inspect eval approval.py --approval $APPROVAL_YAML --model openai/gpt-4o-mini" C-m
done

# Attach to the tmux session
tmux attach-session -t $SESSION_NAME
