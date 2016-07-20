#!/bin/sh

find vendor -type d -name views | while read dir; do
  install -d "$dir" "$1/$dir"
  cp -a "$dir"/* "$1/$dir"
done
