#!/bin/sh

# A helper script used to setup GOPATH correctly 
# so tools like go linter can work correctly

project="$(dirname $(readlink -f $0))/.."
export WEB_ROOT=$(readlink -f "$project")
export GOPATH=$project/vendor:$project

exec $project/bin/core start
