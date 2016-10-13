#!/bin/bash

cd "../../proto"

cat << EOF
package main

// This file is auto-generated
// See generate-helpers.sh for the reference
// Do not try to edit this manually

EOF

pkgs=""

for pkg in $(find . -name '*.part.proto' | awk -F/ '{print $2}' | sort | uniq); do
  echo "import \"proto/$pkg\""
  pkgs="$pkgs $pkg"
done

cat << EOF
var services map[string]interface{}
func connect() {
  services = map[string]interface{}{

EOF

for pkg in $pkgs; do
  cd "$pkg"

    cat *.part.proto | grep ^service | awk '{print $2}' | while read service; do
cat << EOF
      "$service": $pkg.New${service}Client(conn),
EOF
    done

  cd ..
done

cat << EOF
  }
}
EOF
