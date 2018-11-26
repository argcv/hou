#!/usr/bin/env bash

# Config
ORG_NAME='yuikns'
PROJ_NAME='hou'

if [ -z $GOPATH ]; then
  echo "error: env: GOPATH not exists!!"
  exit 1
fi

ORG_DIR="$GOPATH/src/github.com/$ORG_NAME"
PROJ_DIR="$ORG_DIR/$PROJ_NAME"

if [ -d $PROJ_DIR ]; then
  echo "error: project is exists!!"
  exit 2
fi

git clone git@github.com:$ORG_NAME/$PROJ_NAME.git $PROJ_DIR

echo "cloned to $PROJ_DIR"

