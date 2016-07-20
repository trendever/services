#!/bin/sh

# A helper script used to setup GOPATH correctly 
# so tools like go linter can work correctly

project="$(dirname $(readlink -f $0))/.."
export WEB_ROOT=$project
export GOPATH=$project/vendor:$project

if [ -f $project/bin/core ]; then
  exec $project/bin/core start
else
  exec $project/service start
fi
