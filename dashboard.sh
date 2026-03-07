#!/bin/bash
# nullcal-dashboard.sh - Launches nullcal in a Tmux split-pane layout

if ! command -v tmux &> /dev/null; then
    echo "Error: tmux is not installed."
    echo "Please install tmux to use the dashboard layout."
    exit 1
fi

SESSION_NAME="nullcal-dashboard"

# Check if session already exists
tmux has-session -t $SESSION_NAME 2>/dev/null

if [ $? != 0 ]; then
    # Create the session and run the Calendar view on the left (50%)
    tmux new-session -d -s $SESSION_NAME "./bin/nullcal view week"
    
    # Split horizontally (right pane), creating a Top/Bottom split on the right 50%
    tmux split-window -h -p 50 "./bin/nullcal view todo"
    
    # Split the right pane vertically for the Kanban board (bottom right)
    tmux split-window -v -p 60 "./bin/nullcal view kanban"
    
    # Select the calendar pane
    tmux select-pane -t 0
fi

# Attach to the session
tmux attach-session -t $SESSION_NAME
