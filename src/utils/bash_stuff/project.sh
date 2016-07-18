#!/bin/bash

# A helper script used to setup GOPATH correctly 
# so tools like go linter can work correctly

project="$PWD"
echo "Project $project"
export GOPATH=$project/vendor:$project
unset project
