#!/bin/bash

if [ -n "$1" ]; then
	echo "Usage: $0 (in dir with proto.part files)"
	exit 1
fi


path="$PWD"
pkg=$(basename "${path}")
proto="${path}/${pkg}.proto"

cat - ${path}/*.part.proto > "${proto}" << EOF
syntax = "proto3";
package $pkg;
EOF

protoc --proto_path="$PWD" --go_out=plugins=grpc:. "${proto}"
rm "${proto}"

