#!/bin/bash

ls vendor/manifest || (echo "Manifest not found!"; exit 1)

rm -rf pkg vendor/pkg
cd vendor/

for dep in $(cat manifest | grep importpath | tr '"' '\n' | grep /); do 
  grep -rl $dep ../src/ src/ | grep '.go$' | grep -qv $dep || echo $dep; done | \
  xargs -n1 echo gb vendor delete
