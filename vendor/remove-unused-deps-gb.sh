#!/bin/bash

dir=$(dirname "$(readlink "$0")")
cd $dir
ls manifest || (echo "Manifest not found!"; exit 1)

for dep in $(cat manifest | grep importpath | tr '"' '\n' | grep /); do
  grep -rl $dep ../src/ src/ | grep -qv $dep || echo $dep; done | \
  xargs -n1 echo gb vendor delete
