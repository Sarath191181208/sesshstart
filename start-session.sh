#!/bin/bash 

###### VENV ####### 
# Check if virtual environment is activated
if [ -z "VIRTUAL_ENV" ]
then
    printf "Virtual environment is not activated Use `${BYellow} pipenv ${BIWhite}shell${NC}` command. Exiting..."
    exit 1
fi

###### Docker ####### 
# Check if Docker is running
if ! docker info &> /dev/null
then
    printf "Docker is not running. Exiting..."
    exit 1
fi

SESSIONNAME="sesshstart"

# check for session 
tmux has-session -t SESSIONNAME &> /dev/null
if [ $? != 0 ] 
 then
  # start the new session 
  tmux new-session -d -s SESSIONNAME

  # Window 1: Open nvim on the current project root
  tmux new-window -t SESSIONNAME:1 -n 'nvim'
  tmux send-keys -t SESSIONNAME:1 'nvim' C-m

  # Window 2: Command line window
  tmux new-window -t SESSIONNAME:2 -n 'cmd'

fi

tmux attach -t SESSIONNAME
