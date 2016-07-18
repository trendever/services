#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 directory_name"
  exit 1
fi

find $1 -name 'proto.package' | while read protopkg; do

  path=$(dirname "${protopkg}")
  pkg=$(basename "${path}")
  proto="${path}/${pkg}.proto"

  cat - ${path}/*.part.proto > "${proto}" << EOF
    syntax = "proto3";
    package $pkg;
EOF

  protoc --go_out=plugins=grpc:. "${proto}"
  rm "${proto}"


done
