#!/bin/bash

# Name of the tmux session
SESSION_NAME="my_session"

# Number of tasks
PANE_COUNT=4

# Calculate the midpoint
MIDPOINT=$((PANE_COUNT / 2))

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
    # Set APPROVAL_YAML based on whether we're in the first or second half
    if [ $i -lt $MIDPOINT ]; then
        APPROVAL_YAML="approval_1.yaml"
    else
        APPROVAL_YAML="approval_4.yaml"
    fi

    TASK="${TASKS[$i]}"
    # Escape double quotes in the task string
    ESCAPED_TASK=$(printf '%q' "$TASK")
    tmux send-keys -t $SESSION_NAME.$i "inspect eval approval.py --approval $APPROVAL_YAML --model openai/gpt-4o-mini --trace" C-m
    
    # Print the selected file for debugging
    echo "Pane $i: Selected approval file: $APPROVAL_YAML"
done

# Attach to the tmux session
tmux attach-session -t $SESSION_NAME
